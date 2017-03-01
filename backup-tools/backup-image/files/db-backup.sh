#!/bin/bash

source /env.cron

PG_DUMP=1

EXPIRE=0

SWIFT_CONTAINER="db_backup/${OS_REGION_NAME}/${MY_POD_NAMESPACE}/${MY_POD_NAME}"

# We assume that the databases are using their default ports
MYSQL_PORT=3306
MYSQL_SOCKET=/db/socket/mysqld.sock
PGSQL_PORT=5432
PGSQL_SOCKET=/db/socket/.s.PGSQL.$PGSQL_PORT
PGSQL_BARMAN_DIR=/var/lib/barman/

CUR_TS="$(date +%Y%m%d%H%M)"
LAST_BACKUP_FILE="/tmp/last_backup_timestap"
PIDFILE="/var/run/db-backup.pid"

if [ -f "$LAST_BACKUP_FILE" ] ; then
  LAST_BACKUP_TS="$(cat $LAST_BACKUP_FILE)"
else
  LAST_BACKUP_TS=0
fi

if [ -f "$PIDFILE" ] ; then
  PID="`cat $PIDFILE`"
  if [ -e /proc/$PID -a /proc/$PID/exe ] ; then
    echo "Backup already in progress..."
    exit 1
  fi
fi

if [ "$BACKUP_MYSQL_FULL" ] && [ "$BACKUP_MYSQL_INCR" ] && [ -S $MYSQL_SOCKET ] ; then

  DATADIR=/db/data/
  SOCKET=/db/socket/mysqld.sock
  BACKUP_BASE=/backup/mysql/base/
  USERNAME=root
  PASSWORD=$MYSQL_ROOT_PASSWORD

  if [ ! -d "$BACKUP_BASE" ] ; then
    mkdir -p "$BACKUP_BASE"
  fi

  # Create backup interval
  INTERVAL_FULL="$BACKUP_MYSQL_FULL"
  INTERVAL_INCR="$BACKUP_MYSQL_INCR"
  IS_NEXT_TS_FULL="$(date --date="now - $INTERVAL_FULL" +%Y%m%d%H%M)"
  IS_NEXT_TS_INCR="$(date --date="now - $INTERVAL_INCR" +%Y%m%d%H%M)"

  if [ "$IS_NEXT_TS_FULL" -ge "$LAST_BACKUP_TS" ] ; then
    echo $$ > $PIDFILE
    echo "$CUR_TS" > $LAST_BACKUP_FILE

    # MySQL Backup (full)
    xtrabackup --backup --user=$USERNAME --password=$PASSWORD --target-dir=$BACKUP_BASE --datadir=$DATADIR --socket=$MYSQL_SOCKET
    swift upload "$SWIFT_CONTAINER/base" $BACKUP_BASE

    rm $PIDFILE
    exit 0
  fi

  if [ "$IS_NEXT_TS_INCR" -ge "$LAST_BACKUP_TS" ] ; then
    echo $$ > $PIDFILE
    echo "$CUR_TS" > $LAST_BACKUP_FILE

    if [ -d "/backup/inc$LAST_BACKUP_TS" ] ; then
      BACKUP_BASE=/backup/inc$LAST_BACKUP_TS
    fi

    # MySQL Backup (incremental)
    xtrabackup --backup --user=$USERNAME --password=$PASSWORD --target-dir=/backup/inc$CUR_TS --incremental-basedir=$BACKUP_BASE --datadir=$DATADIR --socket=$MYSQL_SOCKET
    swift upload "$SWIFT_CONTAINER/inc$CUR_TS" /backup/inc$CUR_TS
    rm $PIDFILE
    exit 0
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
  INTERVAL_INCR="$BACKUP_PGSQL_INCR"
  IS_NEXT_TS_FULL="$(date --date="now - $INTERVAL_FULL" +%Y%m%d%H%M)"

  if [ "$IS_NEXT_TS_FULL" -ge "$LAST_BACKUP_TS" ] ; then
    echo $$ > $PIDFILE
    echo "$CUR_TS" > $LAST_BACKUP_FILE

    if [ "$PG_DUMP" = 1 ] ; then
      for i in `psql -q -A -t -c "SELECT datname FROM pg_database" -h localhost -U postgres | grep -E -v "(^template|^postgres$)"` ; do
        echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Creating backup of database $i ..."
        pg_dump -U postgres -h localhost $i --file=$BACKUP_BASE/$i.sql
        gzip -f $BACKUP_BASE/$i.sql
      done
      echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Uploading backup to $SWIFT_CONTAINER/$CUR_TS ..."
      swift upload --changed "$SWIFT_CONTAINER/$CUR_TS" $BACKUP_BASE
    else
      # Postgres Backup (full)
      /usr/bin/barman  cron
      /usr/bin/barman backup all
      swift upload --changed "$SWIFT_CONTAINER/$CUR_TS" $BACKUP_BASE
      swift upload --changed "$SWIFT_CONTAINER/WAL" $PGSQL_BARMAN_DIR
    fi

    rm $PIDFILE
    exit 0
  fi
fi

if [ "$BACKUP_INFLUXDB_FULL" ] ; then
  # InfluxDB Backup
  BACKUP_BASE=/backup/influxdb/base

  if [ ! -d "$BACKUP_BASE" ] ; then
    mkdir -p "$BACKUP_BASE"
  fi

  # Create backup interval
  INTERVAL_FULL="$(cat /etc/db-backup/influxdb.backup.full)"
  IS_NEXT_TS_FULL="$(date --date="now - $INTERVAL_FULL" +%Y%m%d%H%M)"

  if [ "$TESTING" -gt "0" ] ; then
    # Cleanup for testing
    rm -rf /backup/*
  fi

  echo "$IS_NEXT_TS_FULL -ge $LAST_BACKUP_TS"

  if [ "$IS_NEXT_TS_FULL" -ge "$LAST_BACKUP_TS" ] ; then
    echo $$ > $PIDFILE
    echo "$CUR_TS" > $LAST_BACKUP_FILE

    for i in `influx -execute 'show databases' -host localhost:8083 | grep -E -v "(^---|^_internal|^name)"` ; do
      echo "[$(date +%Y%m%d%H%M%S)] Creating backup of database $i ..."
      influxd backup -database $i -host localhost:8083 "$BACKUP_BASE/$i"
      tar zcvf "$i.tar.gz" "$BACKUP_BASE/$i"
      rm -rf $BACKUP_BASE/$i
    done
    echo "[$(date +%Y%m%d%H%M%S)] Uploading backup to influxdb/$SWIFT_CONTAINER/$CUR_TS ..."
    swift upload --changed "$SWIFT_CONTAINER/$CUR_TS" $BACKUP_BASE

    rm $PIDFILE
    exit 0
  fi
fi

if [ "$EXPIRE" = 1 ] ; then
  if [ "$BACKUP_EXPIRATION_INTERVAL" = "" ] ; then
    BACKUP_EXPIRE="10 days"
  else
    BACKUP_EXPIRE="$BACKUP_EXPIRATION_INTERVAL"
  fi
  EXPIRE_DATE="`date -d -\"$BACKUP_EXPIRE\" +\"%Y%m%d%H%M\"`"
  for i in `swift list $BACKUP_BASE` ; do
    BACKUP_DATE="`echo $i | cut -d / -f 1`"
    if [ "$BACKUP_DATE" -le "$EXPIRE_DATE" ] ; then
      echo "swift delete db_backup $BACKUP_BASE/$i"
    fi
  done
fi
