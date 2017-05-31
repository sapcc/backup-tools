#!/bin/bash

if [ "$BACKUP_EXPIRE_AFTER" = "" ] ; then
  BACKUP_EXPIRE_AFTER=864000
fi

if [ ! -f /backup/env/from.env ] || [ ! -f /backup/env/to0.env ] ; then
  echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Configuration files missing, check helm deployment."
  exit 1
fi

PIDFILE="/var/run/backup-replication.pid"

if [ -f "$PIDFILE" ] ; then
  PID="`cat $PIDFILE`"
  if [ -e /proc/$PID -a /proc/$PID/exe ] ; then
    echo "Replication already in progress..."
    exit 1
  fi
fi

source /backup/env/from.env

REPLICATE_FROM="$OS_REGION_NAME"

if [ ! -d /backup/tmp ] ; then
  mkdir /backup/tmp
fi

cd /backup/tmp

echo $$ > $PIDFILE

source /backup/env/from.env
swift list db_backup --prefix "$REPLICATE_FROM" > /backup/tmp/from.log

for i in /backup/env/to*.env ; do
  source $i
  REPLICATE_TO="$OS_REGION_NAME"

  swift list db_backup --prefix "$REPLICATE_FROM" > /backup/tmp/to.log

  REPL_OBJECTS="`comm --nocheck-order -23 /backup/tmp/from.log /backup/tmp/to.log`"

  if [ "$REPL_OBJECTS" != "" ] ; then
    source /backup/env/from.env

    for j in $REPL_OBJECTS ; do
      if [ ! -f $j ] ; then
        echo -n "$(date +'%Y/%m/%d %H:%M:%S %Z') Downloading from $REPLICATE_FROM: "
        source /backup/env/from.env
        swift download db_backup $j
        echo -n "$(date +'%Y/%m/%d %H:%M:%S %Z') Uploading to $REPLICATE_TO: "
        source $i
        swift upload --header "X-Delete-After: $BACKUP_EXPIRE_AFTER" db_backup $j
        rm -rf $j
      fi
    done
    rm -rf /backup/tmp/to.log
  else
    echo "$(date +'%Y/%m/%d %H:%M:%S %Z') No new backups to transfer."
  fi
done

date +'%s' > /backup/tmp/last_run

rm $PIDFILE
