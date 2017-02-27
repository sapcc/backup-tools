#!/bin/bash

source /backup/env/from.env

LOGFILE="/proc/1/fd/1"

REPLICATE_FROM="$OS_REGION_NAME"
REPLICATE_TO="eu-de-1"

if [ ! -d /backup/tmp ] ; then
  mkdir /backup/tmp
fi

cd /backup/tmp

source /backup/env/from.env
swift list db_backup | grep "^$REPLICATE_FROM/" > /backup/tmp/from.log

source /backup/env/to1.env
swift list db_backup | grep "^$REPLICATE_FROM/" > /backup/tmp/to.log

REPL_OBJECTS="`cat /backup/tmp/from.log /backup/tmp/to.log | sort | uniq -u`"

if [ "$REPL_OBJECTS" != "" ] ; then
  source /backup/env/from.env

  echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Downloading backups from $REPLICATE_FROM..." > $LOGFILE
  for i in $REPL_OBJECTS ; do
    echo -n "$(date +'%Y/%m/%d %H:%M:%S %Z') " > $LOGFILE
    swift download db_backup $i > $LOGFILE
  done

  source /backup/env/to1.env

  echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Uploading backups to $REPLICATE_TO..." > $LOGFILE
  for i in $REPL_OBJECTS ; do
    echo -n "$(date +'%Y/%m/%d %H:%M:%S %Z') " > $LOGFILE
    swift upload db_backup $i > $LOGFILE
  done

  rm -rf /backup/tmp/*
else
  echo "$(date +'%Y/%m/%d %H:%M:%S %Z') No new backups to transfer." > $LOGFILE
fi
