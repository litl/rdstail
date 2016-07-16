#!/bin/sh

if [ -z "$RDSHOST" ] || [ -z "$POLLING_RATE" ] || [ -z "$PAPERTRAIL_HOST" ] || [ -z "$PROGRAM" ]; then
  echo "required env vars not set"
  exit 1
fi

/app/rdstail -i $RDSHOST \
  papertrail -p $PAPERTRAIL_HOST \
  -a $PROGRAM \
  --hostname $RDSHOST \
  -r $POLLING_RATE
