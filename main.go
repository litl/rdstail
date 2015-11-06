package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/codegangsta/cli"
	"github.com/litl/rdstail/src"
)

func signalListen(stop chan<- struct{}) {
	c := make(chan os.Signal)
	signal.Notify(c, syscall.SIGTERM, syscall.SIGINT)

	<-c
	close(stop)
	<-c
	log.Panic("Aborting on second signal")
}

func run(c *cli.Context) {
	r := rds.New(session.New(), aws.NewConfig().WithRegion(c.String("region")))
	db := c.String("instance")
	numLines := int64(c.Int("lines"))
	papertrailHost := c.String("papertrail")
	appName := c.String("app")
	rate, err := time.ParseDuration(c.String("rate"))
	if err != nil {
		log.Fatal(err)
	}

	stop := make(chan struct{})
	go signalListen(stop)

	switch {
	case c.Bool("watch"):
		err = rdstail.Watch(r, db, rate, func(lines string) error {
			fmt.Print(lines)
			return nil
		}, stop)
	case papertrailHost != "":
		err = rdstail.FeedPapertrail(r, db, rate, papertrailHost, appName, stop)
	default:
		err = rdstail.Tail(r, db, numLines)
	}

	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	app := cli.NewApp()

	app.Name = "rdstail"
	app.Usage = "Prints AWS RDS logs"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.IntFlag{
			Name:  "lines, n",
			Value: 20,
			Usage: "output the last n lines. use 0 for a full dump of the most recent file",
		},
		cli.StringFlag{
			Name:  "instance, i",
			Value: "dev",
			Usage: "name of the db instance in rds",
		},
		cli.StringFlag{
			Name:   "region",
			Value:  "us-east-1",
			Usage:  "AWS region",
			EnvVar: "AWS_REGION",
		},
		cli.StringFlag{
			Name:  "rate, r",
			Value: "3s",
			Usage: "watch polling rate",
		},
		cli.BoolFlag{
			Name:  "watch, w",
			Usage: "watch mode",
		},
		cli.StringFlag{
			Name:  "papertrail",
			Usage: "if set, streams logs to papertrail over TLS at this host",
		},
		cli.StringFlag{
			Name:  "app, a",
			Value: "rdstail",
			Usage: "app name to send to papertrail",
		},
	}

	app.Run(os.Args)
}
