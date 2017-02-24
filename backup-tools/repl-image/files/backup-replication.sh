#!/bin/bash

source /env_staging.cron

LOGFILE="/proc/1/fd/1"

REPLICATE_FROM="$OS_REGION_NAME"
REPLICATE_TO="eu-de-1"

cd /backup

source /env_staging.cron
swift list db_backup | grep "^$REPLICATE_FROM/" > /backup/from.log

source /env_$REPLICATE_TO.cron
swift list db_backup | grep "^$REPLICATE_FROM/" > /backup/to.log

REPL_OBJECTS="`cat /backup/from.log /backup/to.log | sort | uniq -u`"

if [ "$REPL_OBJECTS" != "" ] ; then
  source /env_$REPLICATE_FROM.cron

  echo "$(date +'%Y/%m/%d %H:%M:%S') Downloading backups from $REPLICATE_FROM..." > $LOGFILE
  for i in $REPL_OBJECTS ; do
    echo -n "$(date +'%Y/%m/%d %H:%M:%S') " > $LOGFILE
    swift download db_backup $i > $LOGFILE
  done

  source /env_$REPLICATE_TO.cron

  echo "$(date +'%Y/%m/%d %H:%M:%S') Uploading backups to $REPLICATE_TO..." > $LOGFILE
  for i in $REPL_OBJECTS ; do
    echo -n "$(date +'%Y/%m/%d %H:%M:%S') " > $LOGFILE
    swift upload db_backup $i > $LOGFILE
  done

  rm -rf /backup/*
else
  echo "$(date +'%Y/%m/%d %H:%M:%S') No new backups to transfer." > $LOGFILE
fi
