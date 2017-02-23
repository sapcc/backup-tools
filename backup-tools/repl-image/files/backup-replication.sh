#!/bin/bash

LOGFILE="/proc/1/fd/1"

FROM="staging"
TO="eu-de-1"

cd /backup

source /env_$FROM.cron
swift list db_backup | grep "^$FROM/" > /backup/from.log

source /env_$TO.cron
swift list db_replication | grep "^$FROM/" > /backup/to.log

REPL_OBJECTS="`cat /backup/from.log /backup/to.log | sort | uniq -u`"

source /env_$FROM.cron

echo "[$(date +%Y%m%d%H%M%S)] Downloading backups from $FROM..." > $LOGFILE
for i in $REPL_OBJECTS ; do
  echo -n "[$(date +%Y%m%d%H%M%S)] " > $LOGFILE
  swift download db_backup $i > $LOGFILE
done

source /env_$TO.cron

echo "[$(date +%Y%m%d%H%M%S)] Uploading backups to $TO..." > $LOGFILE
for i in $REPL_OBJECTS ; do
  echo -n "[$(date +%Y%m%d%H%M%S)] " > $LOGFILE
  swift upload db_replication $i > $LOGFILE
done

rm -rf /backup/*
