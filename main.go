package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/litl/rdstail/src"
	"github.com/urfave/cli"
)

func fie(e error) {
	if e != nil {
		fmt.Println(e)
		os.Exit(1)
	}
}

func signalListen(stop chan<- struct{}) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	<-c
	close(stop)
	<-c
	log.Panic("Aborting on second signal")
}

func setupRDS(c *cli.Context) *rds.RDS {
	region := c.GlobalString("region")
	maxRetries := c.GlobalInt("max-retries")
	cfg := aws.NewConfig().WithRegion(region).WithMaxRetries(maxRetries)
	return rds.New(session.New(), cfg)
}

func parseRate(c *cli.Context) time.Duration {
	rate, err := time.ParseDuration(c.String("rate"))
	fie(err)
	return rate
}

func parseDB(c *cli.Context) string {
	db := c.GlobalString("instance")
	if db == "" {
		fie(errors.New("-instance required"))
	}
	return db
}

func parseFile(c *cli.Context) string {
	file := c.String("file")
	return file
}

func watch(c *cli.Context) {
	r := setupRDS(c)
	db := parseDB(c)
	rate := parseRate(c)
	file := parseFile(c)
	stop := make(chan struct{})
	go signalListen(stop)

	if file != "" {
		err := rdstail.WatchFile(r, db, rate, file, func(lines string) error {
			fmt.Print(lines)
			return nil
		}, stop)

		fie(err)
	}

	err := rdstail.Watch(r, db, rate, func(lines string) error {
		fmt.Print(lines)
		return nil
	}, stop)

	fie(err)
}

func papertrail(c *cli.Context) {
	r := setupRDS(c)
	db := parseDB(c)
	rate := parseRate(c)
	papertrailHost := c.String("papertrail")
	if papertrailHost == "" {
		fie(errors.New("-papertrail required"))
	}
	appName := c.String("app")
	hostname := c.String("hostname")
	if hostname == "os.Hostname()" {
		var err error
		hostname, err = os.Hostname()
		fie(err)
	}

	stop := make(chan struct{})
	go signalListen(stop)

	err := rdstail.FeedPapertrail(r, db, rate, papertrailHost, appName, hostname, stop)

	fie(err)
}

func tail(c *cli.Context) {
	r := setupRDS(c)
	db := parseDB(c)
	file := parseFile(c)
	numLines := int64(c.Int("lines"))
	if file != "" {
		err := rdstail.TailFile(r, db, file, numLines)
		fie(err)
	}
	err := rdstail.Tail(r, db, numLines)
	fie(err)
}

func main() {
	app := cli.NewApp()

	app.Name = "rdstail"
	app.Usage = `Reads AWS RDS logs

    AWS credentials are taken from an ~/.aws/credentials file or the env vars AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.`
	app.Version = "0.1.0"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "instance, i",
			Usage: "name of the db instance in rds [required]",
		},
		cli.StringFlag{
			Name:   "region",
			Value:  "us-east-1",
			Usage:  "AWS region",
			EnvVar: "AWS_REGION",
		},
		cli.IntFlag{
			Name:  "max-retries",
			Value: 10,
			Usage: "maximium number of retries for rds requests",
		},
	}

	app.Commands = []cli.Command{
		{
			Name:   "papertrail",
			Usage:  "stream logs into papertrail",
			Action: papertrail,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "papertrail, p",
					Value: "",
					Usage: "papertrail host e.g. logs.papertrailapp.com:8888 [required]",
				},
				cli.StringFlag{
					Name:  "app, a",
					Value: "rdstail",
					Usage: "app name to send to papertrail",
				},
				cli.StringFlag{
					Name:  "hostname",
					Value: "os.Hostname()",
					Usage: "hostname of the client, sent to papertrail",
				},
				cli.StringFlag{
					Name:  "rate, r",
					Value: "3s",
					Usage: "rds log polling rate",
				},
			},
		},

		{
			Name:   "watch",
			Usage:  "stream logs to stdout",
			Action: watch,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "rate, r",
					Value: "3s",
					Usage: "rds log polling rate",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "name of the logfile to watch",
				},
			},
		},

		{
			Name:   "tail",
			Usage:  "tail the last N lines",
			Action: tail,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "lines, n",
					Value: 20,
					Usage: "output the last n lines. use 0 for a full dump of the most recent file",
				},
				cli.StringFlag{
					Name:  "file, f",
					Usage: "name of the logfile to tail",
				},
			},
		},
	}

	app.Run(os.Args)
}
