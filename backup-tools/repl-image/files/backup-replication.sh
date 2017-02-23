#!/bin/bash

source /env.cron

LOGFILE="/proc/1/fd/1"
REPLICATE_TO="eu-de-1"

cd /backup

source /env.cron
swift list db_backup | grep "^$OS_REGION_NAME/" > /backup/from.log

source /env_$REPLICATE_TO.cron
swift list db_replication | grep "^$OS_REGION_NAME/" > /backup/to.log

REPL_OBJECTS="`cat /backup/from.log /backup/to.log | sort | uniq -u`"

source /env_$OS_REGION_NAME.cron

echo "[$(date +%Y%m%d%H%M%S)] Downloading backups from $OS_REGION_NAME..." > $LOGFILE
for i in $REPL_OBJECTS ; do
  echo -n "[$(date +%Y%m%d%H%M%S)] " > $LOGFILE
  swift download db_backup $i > $LOGFILE
done

source /env_$REPLICATE_TO.cron

echo "[$(date +%Y%m%d%H%M%S)] Uploading backups to $TO..." > $LOGFILE
for i in $REPL_OBJECTS ; do
  echo -n "[$(date +%Y%m%d%H%M%S)] " > $LOGFILE
  swift upload db_replication $i > $LOGFILE
done

rm -rf /backup/*
