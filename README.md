RDSTail
=======

RDSTail is a tool for tailing or streaming RDS log files.  Supports piping to papertrail.

Installation
============

For now, you must compile from source.  Install [https://golang.org](Go).

    » go get github.com/litl/rdstail


Usage
=====

```
» ./rdstail -h

NAME:
   rdstail - Reads AWS RDS logs

    AWS credentials are taken from an ~/.aws/credentials file or the env vars AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY.

USAGE:
   ./rdstail [global options] command [command options] [arguments...]
   
VERSION:
   0.1.0
   
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
   
------------------------------------------------------------
» ./rdstail papertrail -h

NAME:
   ./rdstail papertrail - stream logs into papertrail

USAGE:
   ./rdstail papertrail [command options] [arguments...]

OPTIONS:
   --papertrail, -p         papertrail host e.g. logs.papertrailapp.com:8888 [required]
   --app, -a "rdstail"      app name to send to papertrail
   --hostname "os.Hostname()"   hostname of the client, sent to papertrail
   --rate, -r "3s"      rds log polling rate
   
------------------------------------------------------------
» ./rdstail watch -h

NAME:
   ./rdstail watch - stream logs to stdout

USAGE:
   ./rdstail watch [command options] [arguments...]

OPTIONS:
   --rate, -r "3s"  rds log polling rate
   
------------------------------------------------------------
» ./rdstail tail -h

NAME:
   ./rdstail tail - tail the last N lines

USAGE:
   ./rdstail tail [command options] [arguments...]

OPTIONS:
   --lines, -n "20" output the last n lines. use 0 for a full dump of the most recent file

```
