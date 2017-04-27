package lib

import (
	"bytes"
	"crypto/tls"
	"errors"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/chrismrivera/backoff"
)

const (
	papertrailBackoffMaxWait  = time.Minute
	papertrailBackoffDeadline = time.Minute * 5
	// aws-sdk-go already offers retry functionality
)

func getMostRecentLogFile(r *rds.RDS, db string) (file *rds.DescribeDBLogFilesDetails, err error) {
	yesterday := time.Now().Add(-24 * time.Hour).Unix()
	file, err = getMostRecentLogFileSince(r, db, yesterday)
	if err != nil {
		return
	}

	if file == nil {
		lastWeek := time.Now().Add(-7 * 24 * time.Hour).Unix()
		file, err = getMostRecentLogFileSince(r, db, lastWeek)
		if err != nil {
			return
		}
	}

	if file == nil {
		file, err = getMostRecentLogFileSince(r, db, 0)
		if err != nil {
			return
		}
	}

	return
}

// getMostRecentLogFileSince returns the most recent log file whose lastWritten >= since.
func getMostRecentLogFileSince(r *rds.RDS, db string, since int64) (file *rds.DescribeDBLogFilesDetails, err error) {
	resp, err := describeLogFiles(r, db, since)
	if err != nil {
		return nil, err
	}
	for _, d := range resp {
		hasData := d.LastWritten != nil && d.LogFileName != nil
		isNewer := file == nil || file.LastWritten == nil || *d.LastWritten > *file.LastWritten
		if hasData && isNewer {
			file = d
		}
	}
	return
}

func describeLogFiles(r *rds.RDS, db string, since int64) (details []*rds.DescribeDBLogFilesDetails, err error) {
	req := &rds.DescribeDBLogFilesInput{
		DBInstanceIdentifier: aws.String(db),
	}
	if since != 0 {
		req.FileLastWritten = aws.Int64(since)
	}

	err = r.DescribeDBLogFilesPages(req, func(p *rds.DescribeDBLogFilesOutput, lastPage bool) bool {
		details = append(details, p.DescribeDBLogFiles...)
		return true
	})

	return
}

func tailLogFile(r *rds.RDS, db, name string, numLines int64, marker string) (string, string, error) {
	req := &rds.DownloadDBLogFilePortionInput{
		DBInstanceIdentifier: aws.String(db),
		LogFileName:          aws.String(name),
	}
	if numLines != 0 {
		req.NumberOfLines = aws.Int64(numLines)
	}
	if marker != "" {
		req.Marker = aws.String(marker)
	}

	var buf bytes.Buffer
	var markerPtr *string
	err := r.DownloadDBLogFilePortionPages(req, func(p *rds.DownloadDBLogFilePortionOutput, lastPage bool) bool {
		if p.LogFileData != nil {
			buf.WriteString(*p.LogFileData)
		}
		if lastPage {
			markerPtr = p.Marker
		}
		return true
	})

	marker = ""
	if markerPtr != nil {
		marker = *markerPtr
	}

	return buf.String(), marker, err
}

/// cmds

func Tail(r *rds.RDS, db string, numLines int64) error {
	logFile, err := getMostRecentLogFile(r, db)
	if err != nil {
		return nil
	}
	if logFile == nil {
		return errors.New("no log file found")
	}

	tail, _, err := tailLogFile(r, db, *logFile.LogFileName, numLines, "")
	if err != nil {
		return err
	}
	fmt.Println(tail)
	return nil
}

func Watch(r *rds.RDS, db string, rate time.Duration, callback func(string) error, stop <-chan struct{}) error {
	// Periodically check for new log files (unless there is a way to detect the file is done being written to)
	// Poll that log file, retaining the marker
	logFile, err := getMostRecentLogFile(r, db)
	if err != nil {
		return err
	}
	if logFile == nil {
		return errors.New("No log files found for this RDS instance.")
	}

	// Get a marker for the end of the log file by requesting the most recent line
	lines, marker, err := tailLogFile(r, db, *logFile.LogFileName, 1, "")
	if err != nil {
		return err
	}

	log.Infof("Starting with most recent log file: %s", *logFile.LogFileName)

	t := time.NewTicker(rate)
	empty := 0
	const checkLogfileRate = 4
	for {
		select {
		case <-t.C:
			// If the logfile tail was empty n times, check for a newer log file
			if empty >= checkLogfileRate {
				log.Debugf("Checking for a new log file since no new logs are found in last %v",
					checkLogfileRate*rate)
				empty = 0
				newLogFile, err := getMostRecentLogFileSince(r, db, *logFile.LastWritten)
				if err != nil {
					return err
				}

				// Ensure if we got a real new log file, and not the same file we are
				// already tailing. Reset the marker if its a real new log file only.
				if newLogFile != nil && *newLogFile.LogFileName != *logFile.LogFileName {
					log.Infof("Found a new log file: %s", *newLogFile.LogFileName)
					logFile = newLogFile
					marker = ""
				}
			}

			lines, marker, err = tailLogFile(r, db, *logFile.LogFileName, 0, marker)
			if err != nil {
				return err
			}

			if lines == "" {
				empty++
			} else {
				empty = 0
				if err := callback(lines); err != nil {
					return err
				}
			}
		case <-stop:
			return nil
		}
	}
}

func FeedPapertrail(r *rds.RDS, db string, rate time.Duration, papertrailHost, app, hostname string, stop <-chan struct{}) error {
	nameSegment := fmt.Sprintf(" %s %s: ", hostname, app)

	// Establish TLS connection with papertrail
	conn, err := tls.Dial("tcp", papertrailHost, &tls.Config{})
	if err != nil {
		return err
	}
	defer conn.Close()

	// watch with callback writing to the connection
	return Watch(r, db, rate, func(lines string) error {
		timestamp := time.Now().UTC().Format("2006-01-02T15:04:05")
		buf := bytes.Buffer{}
		buf.WriteString(timestamp)
		buf.WriteString(nameSegment)
		buf.WriteString(lines)
		return backoff.Try(papertrailBackoffMaxWait, papertrailBackoffDeadline, func() error {
			_, err := conn.Write(buf.Bytes())
			if err != nil {
				log.Warnf("Writing to Papertrail failed. Error: %v. This will be retried.",
					err, papertrailBackoffMaxWait)
			}
			return err
		})
	}, stop)
}
