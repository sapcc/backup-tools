#!/bin/bash

SWIFT_CONTAINER="db_backup"
SWIFT_PREFIX="${OS_REGION_NAME}/${MY_POD_NAMESPACE}/${MY_POD_NAME}"

SEGMENT_SIZE=2147483648

: "${BACKUP_EXPIRE_AFTER:=864000}"
: "${PGSQL_HOST:=localhost}" # NOTE: port 5432 is implied
: "${PGSQL_USER:=postgres}"

CUR_TS="$(date +%Y%m%d%H%M)"
LAST_BACKUP_FILE="last_backup_timestamp"
PIDFILE="/var/run/db-backup.pid"

echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Downloading last backup timestamp from $SWIFT_CONTAINER/ ..."
timeout 3m swift download -o /tmp/$LAST_BACKUP_FILE db_backup $SWIFT_PREFIX/$LAST_BACKUP_FILE

if [ -f "/tmp/$LAST_BACKUP_FILE" ] ; then
  LAST_BACKUP_TS="$(cat /tmp/$LAST_BACKUP_FILE)"
else
  LAST_BACKUP_TS=0
fi

if [ -f "$PIDFILE" ] ; then
  PID="`cat $PIDFILE`"
  if [ -e /proc/$PID -a -e /proc/$PID/exe ] ; then
    echo "Backup already in progress..."
    exit 1
  fi
fi

if [ "$BACKUP_PGSQL_FULL" ] ; then
  # PostgreSQL Backup
  DATADIR=/db/data
  BACKUP_BASE=/backup/pgsql/base

  if [ ! -d "$BACKUP_BASE" ] ; then
    mkdir -p "$BACKUP_BASE"
  fi

  # Create backup interval
  INTERVAL_FULL="$BACKUP_PGSQL_FULL"
  IS_NEXT_TS_FULL="$(date --date="now - $INTERVAL_FULL" +%Y%m%d%H%M)"

  if [ "$IS_NEXT_TS_FULL" -ge "$LAST_BACKUP_TS" ] ; then
    echo $$ > $PIDFILE

    for i in `psql -q -A -t -c "SELECT datname FROM pg_database" -h $PGSQL_HOST -U $PGSQL_USER | grep -E -v "(^template|^postgres$)"` ; do
      echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Creating backup of database $i ..."
      pg_dump -U $PGSQL_USER -h $PGSQL_HOST -c --if-exist -C $i --file=$BACKUP_BASE/$i.sql.gz -Z 5 || exit 1
      if [ -s "$BACKUP_BASE/$i.sql.gz" ] ; then
        echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Uploading backup to $SWIFT_CONTAINER/$CUR_TS ..."
        swift upload --segment-size $SEGMENT_SIZE --use-slo --header "X-Delete-After: $BACKUP_EXPIRE_AFTER" --changed --object-name "$SWIFT_PREFIX/$CUR_TS$BACKUP_BASE/$i.sql.gz" "$SWIFT_CONTAINER" "$BACKUP_BASE/$i.sql.gz" || exit 1
      fi
    done

    echo "$CUR_TS" > /tmp/$LAST_BACKUP_FILE
    swift upload --object-name "$SWIFT_PREFIX/$LAST_BACKUP_FILE" "$SWIFT_CONTAINER" "/tmp/$LAST_BACKUP_FILE"
  fi
fi

rm -f $PIDFILE
exit 0
