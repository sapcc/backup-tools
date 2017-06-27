#!/bin/bash

source /env.cron

PG_DUMP=1

SWIFT_CONTAINER="db_backup/${OS_REGION_NAME}/${MY_POD_NAMESPACE}/${MY_POD_NAME}"

if [ "$BACKUP_EXPIRE_AFTER" = "" ] ; then
  BACKUP_EXPIRE_AFTER=864000
fi

# We assume that the databases are using their default ports
MYSQL_PORT=3306
MYSQL_SOCKET=/db/socket/mysqld.sock
PGSQL_PORT=5432
PGSQL_SOCKET=/db/socket/.s.PGSQL.$PGSQL_PORT
PGSQL_BARMAN_DIR=/var/lib/barman/

CUR_TS="$(date +%Y%m%d%H%M)"
LAST_BACKUP_FILE="/tmp/last_backup_timestamp"
PIDFILE="/var/run/db-backup.pid"

swift download -o $LAST_BACKUP_FILE db_backup $OS_REGION_NAME/$MY_POD_NAMESPACE/$MY_POD_NAME$LAST_BACKUP_FILE$LAST_BACKUP_FILE

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

  if [ "$IS_NEXT_TS_FULL" -ge "$LAST_BACKUP_TS" ] || [ "$IS_NEXT_TS_INCR" -ge "$LAST_BACKUP_TS" ] ; then
    echo $$ > $PIDFILE
    echo "$CUR_TS" > $LAST_BACKUP_FILE

    # MySQL Backup (full)
    #xtrabackup --backup --user=$USERNAME --password=$PASSWORD --target-dir=$BACKUP_BASE --datadir=$DATADIR --socket=$MYSQL_SOCKET
    for i in `mysql --user=$USERNAME --password=$PASSWORD --socket=$MYSQL_SOCKET -e 'show databases' | awk '{print $1}' | grep -E -v "(^Database$|^information_schema$|^sys$|^mysql$|^performance_schema$)"`; do
      echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Creating backup of database $i ..."
      mysqldump --opt --user=$USERNAME --password=$PASSWORD --socket=$MYSQL_SOCKET --databases $i > $BACKUP_BASE/$i.sql
      gzip -f $BACKUP_BASE/$i.sql
    done
    echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Uploading backup to $SWIFT_CONTAINER/$CUR_TS ..."
    swift upload --header "X-Delete-After: $BACKUP_EXPIRE_AFTER" --changed "$SWIFT_CONTAINER/$CUR_TS" $BACKUP_BASE

    swift upload $SWIFT_CONTAINER$LAST_BACKUP_FILE $LAST_BACKUP_FILE
    rm -f $LAST_BACKUP_FILE
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
        pg_dump -U postgres -h localhost -c --if-exist -C $i --file=$BACKUP_BASE/$i.sql.gz -Z 5
      done
      if [ -s "$SWIFT_CONTAINER/$CUR_TS" ] ; then
        echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Uploading backup to $SWIFT_CONTAINER/$CUR_TS ..."
        swift upload --header "X-Delete-After: $BACKUP_EXPIRE_AFTER" --changed "$SWIFT_CONTAINER/$CUR_TS" $BACKUP_BASE
      fi
    else
      # Postgres Backup (full)
      /usr/bin/barman  cron
      /usr/bin/barman backup all
      swift upload --header "X-Delete-After: $BACKUP_EXPIRE_AFTER" --changed "$SWIFT_CONTAINER/$CUR_TS" $BACKUP_BASE
      swift upload --header "X-Delete-After: $BACKUP_EXPIRE_AFTER" --changed "$SWIFT_CONTAINER/WAL" $PGSQL_BARMAN_DIR
    fi

    swift upload $SWIFT_CONTAINER$LAST_BACKUP_FILE $LAST_BACKUP_FILE
    rm -f $LAST_BACKUP_FILE
  fi
fi

#if [ "$BACKUP_INFLUXDB_FULL" ] ; then
#  BACKUP_BASE=/backup/influxdb/base
#  USERNAME=root
#  PASSWORD=$INFLUXDB_ROOT_PASSWORD
#
#  if [ ! -d "$BACKUP_BASE" ] ; then
#    mkdir -p "$BACKUP_BASE"
#  fi
#
#  INTERVAL_FULL="$BACKUP_INFLUXDB_FULL"
#  IS_NEXT_TS_FULL="$(date --date="now - $INTERVAL_FULL" +%Y%m%d%H%M)"
#
#  if [ "$IS_NEXT_TS_FULL" -ge "$LAST_BACKUP_TS" ] ; then
#    echo $$ > $PIDFILE
#    echo "$CUR_TS" > $LAST_BACKUP_FILE
#
#    for i in `influx -username "$USERNAME" -password "$PASSWORD" -execute 'show databases' -host localhost -port 8086 | grep -E -v "(^---|^_internal|^name|^$)"` ; do
#      echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Creating backup of database $i ..."
#      influxd backup -database $i -host localhost:8088 "$BACKUP_BASE/$i"
#      tar zcvf "$BACKUP_BASE/$i.tar.gz" "$BACKUP_BASE/$i"
#      rm -rf $BACKUP_BASE/$i
#    done
#    echo "$(date +'%Y/%m/%d %H:%M:%S %Z') Uploading backup to influxdb/$SWIFT_CONTAINER/$CUR_TS ..."
#    swift upload --header "X-Delete-After: $BACKUP_EXPIRE_AFTER" --changed "$SWIFT_CONTAINER/$CUR_TS" $BACKUP_BASE
#
#    swift upload $SWIFT_CONTAINER$LAST_BACKUP_FILE $LAST_BACKUP_FILE
#    rm -f $LAST_BACKUP_FILE
#  fi
#fi

rm -f $PIDFILE
exit 0
