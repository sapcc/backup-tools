#!/bin/bash

if [ "$BACKUP_EXPIRE_AFTER" = "" ] ; then
  BACKUP_EXPIRE_AFTER=864000
fi

if [ ! -f /backup/env/from.env ] || [ ! -f /backup/env/to1.env ] ; then
  echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Configuration files missing, check helm deployment."
  exit 1
fi

PIDFILE="/var/run/backup-replication.pid"

if [ -f "$PIDFILE" ] ; then
  PID="`cat $PIDFILE`"
  if [ -e /proc/$PID -a /proc/$PID/exe ] ; then
    echo "Backup already in progress..."
    exit 1
  fi
fi

source /backup/env/from.env

REPLICATE_FROM="$OS_REGION_NAME"
REPLICATE_TO="eu-de-1"

if [ ! -d /backup/tmp ] ; then
  mkdir /backup/tmp
fi

cd /backup/tmp

echo $$ > $PIDFILE

for i in /backup/env/to*.env ; do

  source /backup/env/from.env
  swift list db_backup | grep "^$REPLICATE_FROM/" > /backup/tmp/from.log

  source $i
  swift list db_backup | grep "^$REPLICATE_FROM/" > /backup/tmp/to.log

  REPL_OBJECTS="`cat /backup/tmp/from.log /backup/tmp/to.log | sort | uniq -u`"

  if [ "$REPL_OBJECTS" != "" ] ; then
    source /backup/env/from.env

    echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Downloading backups from $REPLICATE_FROM..."
    for i in $REPL_OBJECTS ; do
      if [ ! -f $i ] ; then
        echo -n "$(date +'%Y/%m/%d %H:%M:%S %Z') "
        swift download db_backup $i
      fi
    done

    source $i

    echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Uploading backups to $REPLICATE_TO..."
    for i in $REPL_OBJECTS ; do
      echo -n "$(date +'%Y/%m/%d %H:%M:%S %Z') "
      swift upload --header "X-Delete-After: $BACKUP_EXPIRE_AFTER" db_backup $i
    done

    rm -rf /backup/tmp/*
  else
    echo "$(date +'%Y/%m/%d %H:%M:%S %Z') No new backups to transfer."
  fi
done

rm $PIDFILE
