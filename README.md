# RDSTail

RDSTail is a tool for tailing or streaming RDS log files. Supports piping to papertrail. Forked from [litl/rdstail](https://github.com/litl/rdstail).

[![Build Status](https://travis-ci.org/pperzyna/rdstail.svg?branch=master)](https://travis-ci.org/pperzyna/rdstail)
[![Docker Pulls](https://img.shields.io/docker/pulls/pperzyna/rdstail.svg)](https://hub.docker.com/r/pperzyna/rdstail/tags)

## Installation

For now, you must compile from source. Install [Go](https://golang.org).

``` bash
go get github.com/pperzyna/rdstail
cd ${GOPATH}/src/github.com/pperzyna/rdstail/
go build -o rdstail

export AWS_ACCESS_KEY_ID=""
export AWS_SECRET_ACCESS_KEY=""

./rdstail <flags>
```

## Quick Start

This package is available with Docker:

1. Run RDSTail

```bash
docker run -e AWS_ACCESS_KEY_ID="" -e AWS_SECRET_ACCESS_KEY="" pperzyna/rdstail
```

## Usage

``` bash
./rdstail -h

NAME:
    rdstail - Reads AWS RDS logs

    AWS credentials are taken from an ~/.aws/credentials file or the env vars AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.

USAGE:
   ./rdstail [global options] command [command options] [arguments...]
      
COMMANDS:
   papertrail   stream logs into papertrail
   watch    stream logs to stdout
   tail     tail the last N lines
   help, h  Shows a list of commands or help for one command
   
GLOBAL OPTIONS:
   --instance, -i   name of the db instance in rds [required]
   --region "us-east-1" AWS region [$AWS_REGION]
   --max-retries "10"   maximium number of retries for rds requests
   --help, -h       show help
   --version, -v    print the version
   
```
``` bash
./rdstail papertrail -h

NAME:
   ./rdstail papertrail - stream logs into papertrail

USAGE:
   ./rdstail papertrail [command options] [arguments...]

OPTIONS:
   --papertrail, -p         papertrail host e.g. logs.papertrailapp.com:8888 [required]
   --app, -a "rdstail"      app name to send to papertrail
   --hostname "os.Hostname()"   hostname of the client, sent to papertrail
   --rate, -r "3s"      rds log polling rate
```
``` bash
./rdstail watch -h

NAME:
   ./rdstail watch - stream logs to stdout

USAGE:
   ./rdstail watch [command options] [arguments...]

OPTIONS:
   --rate, -r "3s"  rds log polling rate
   --file, -f "trace/alert_DATABASE.log.2017-09-27" name of the logfile
   
```
``` bash
./rdstail tail -h

NAME:
   ./rdstail tail - tail the last N lines

USAGE:
   ./rdstail tail [command options] [arguments...]

OPTIONS:
   --lines, -n "20" output the last n lines. use 0 for a full dump of the most recent file
   --file, -f "trace/alert_DATABASE.log.2017-09-27" name of the logfile
```