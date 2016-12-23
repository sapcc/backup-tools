#!/bin/bash

source /env.cron

TESTING=0
PG_DUMP=1

SWIFT_CONTAINER="${OS_REGION_NAME}_${MY_POD_NAMESPACE}_${MY_POD_NAME}"

# We assume that the databases are using their default ports
MYSQL_PORT=3306
MYSQL_SOCKET=/db/socket/mysqld.sock
PGSQL_PORT=5432
PGSQL_SOCKET=/db/socket/.s.PGSQL.$PGSQL_PORT
PGSQL_BARMAN_DIR=/var/lib/barman/

CUR_TS="$(date +%Y%m%d%H%M)"
LAST_BACKUP_FILE="/tmp/last_backup_timestap"
LOCKFILE="/tmp/backup.lock"

if [ -f "$LAST_BACKUP_FILE" ] ; then
  LAST_BACKUP_TS="$(cat $LAST_BACKUP_FILE)"
else
  LAST_BACKUP_TS=0
fi

if [ -f "$LOCKFILE" ] ; then
  echo "Backup in progress..."
  exit 0
fi

if [ "$BACKUP_MYSQL_FULL" ] && [ "$BACKUP_MYSQL_INCR" ] && [ -S $MYSQL_SOCKET ] ; then

  DATADIR=/db/data/
  SOCKET=/db/socket/mysqld.sock
  BACKUP_BASE=/backup/MySQL/base/
  USERNAME=root
  PASSWORD=foobar

  # Create backup interval
  INTERVAL_FULL="$BACKUP_MYSQL_FULL"
  INTERVAL_INCR="$BACKUP_MYSQL_INCR"
  IS_NEXT_TS_FULL="$(date --date="now - $INTERVAL_FULL" +%Y%m%d%H%M)"
  IS_NEXT_TS_INCR="$(date --date="now - $INTERVAL_INCR" +%Y%m%d%H%M)"

  if [ "$TESTING" -gt "0" ] ; then
    # Cleanup for testing
    rm -rf /backup/*
  fi

  echo "$IS_NEXT_TS_FULL -ge $LAST_BACKUP_TS"
  if [ "$IS_NEXT_TS_FULL" -ge "$LAST_BACKUP_TS" ] ; then
    touch $LOCKFILE
    echo "$CUR_TS" > $LAST_BACKUP_FILE

    # MySQL Backup (full)
    xtrabackup --backup --user=$USERNAME --password=$PASSWORD --target-dir=$BACKUP_BASE --datadir=$DATADIR --socket=$MYSQL_SOCKET
    swift upload "mariadb-$SWIFT_CONTAINER/base" $BACKUP_BASE

    rm $LOCKFILE
    exit 0
  fi

  if [ "$IS_NEXT_TS_INCR" -ge "$LAST_BACKUP_TS" ] ; then
    touch $LOCKFILE
    echo "$CUR_TS" > $LAST_BACKUP_FILE

    if [ -d "/backup/inc$LAST_BACKUP_TS" ] ; then
      BACKUP_BASE=/backup/inc$LAST_BACKUP_TS
    fi

    # MySQL Backup (incremental)

    xtrabackup --backup --user=$USERNAME --password=$PASSWORD --target-dir=/backup/inc$CUR_TS --incremental-basedir=$BACKUP_BASE --datadir=$DATADIR --socket=$MYSQL_SOCKET
    swift upload "mariadb-$SWIFT_CONTAINER/inc$CUR_TS" /backup/inc$CUR_TS
    rm $LOCKFILE
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

  if [ "$TESTING" -gt "0" ] ; then
    # Cleanup for testing
    rm -rf /backup/*
  fi

  echo "$IS_NEXT_TS_FULL -ge $LAST_BACKUP_TS"

  if [ "$IS_NEXT_TS_FULL" -ge "$LAST_BACKUP_TS" ] ; then


    touch $LOCKFILE
    echo "$CUR_TS" > $LAST_BACKUP_FILE

    if [ "$PG_DUMP" = 1 ] ; then
      for i in `psql -q -A -t -c "SELECT datname FROM pg_database" -h localhost -U postgres | grep -E -v "(^template|^postgres$)"` ; do
        echo "[$(date +%Y%m%d%H%M%S)] Creating backup of database $i ..." >> /var/log/backup.log
        pg_dump -U postgres -h localhost $i --file=$BACKUP_BASE/$i.sql
        gzip -f $BACKUP_BASE/$i.sql
      done
      echo "[$(date +%Y%m%d%H%M%S)] Uploading backup to postgres/$SWIFT_CONTAINER/$CUR_TS ..." >> /var/log/backup.log
      swift upload --changed "postgres/$SWIFT_CONTAINER/$CUR_TS" $BACKUP_BASE
    else
      # Postgres Backup (full)
      /usr/bin/barman  cron
      /usr/bin/barman backup all
      swift upload --changed "postgres/$SWIFT_CONTAINER/$CUR_TS" $BACKUP_BASE
      swift upload --changed "postgres/$SWIFT_CONTAINER/WAL" $PGSQL_BARMAN_DIR
    fi

    rm $LOCKFILE
    exit 0
  fi
fi
