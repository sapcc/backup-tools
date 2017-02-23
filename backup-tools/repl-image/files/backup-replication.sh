#!/bin/bash

FROM="staging"
TO="eu-de-1"

cd /backup

source /env_$FROM.cron
swift list db_backup | grep "^$FROM/" > /backup/from.log

source /env_$TO.cron
swift list db_replication | grep "^$FROM/" > /backup/to.log

REPL_OBJECTS="`cat /backup/from.log /backup/to.log | sort | uniq -u`"

source /env_$FROM.cron

for i in $REPL_OBJECTS ; do
  swift download db_backup $i
done

source /env_$TO.cron

for i in $REPL_OBJECTS ; do
  swift upload db_replication $i
done

rm -rf /backup/*
