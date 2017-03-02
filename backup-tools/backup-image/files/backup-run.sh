#!/bin/bash
echo "declare -x MY_POD_NAME=\"${MY_POD_NAME}\"" > /env.cron
echo "declare -x MY_POD_NAMESPACE=\"${MY_POD_NAMESPACE}\"" >> /env.cron
echo "declare -x BACKUP_EXPIRATION_INTERVAL=\"${BACKUP_EXPIRATION_INTERVAL}\"" >> /env.cron
echo "declare -x OS_AUTH_URL=\"${OS_AUTH_URL}\"" >> /env.cron
echo "declare -x OS_AUTH_VERSION=\"${OS_AUTH_VERSION}\"" >> /env.cron
echo "declare -x OS_IDENTITY_API_VERSION=\"${OS_IDENTITY_API_VERSION}\"" >> /env.cron
echo "declare -x OS_USERNAME=\"${OS_USERNAME}\"" >> /env.cron
echo "declare -x OS_USER_DOMAIN_NAME=\"${OS_USER_DOMAIN_NAME}\"" >> /env.cron
echo "declare -x OS_PROJECT_NAME=\"${OS_PROJECT_NAME}\"" >> /env.cron
echo "declare -x OS_PROJECT_DOMAIN_NAME=\"${OS_PROJECT_DOMAIN_NAME}\"" >> /env.cron
echo "declare -x OS_REGION_NAME=\"${OS_REGION_NAME}\"" >> /env.cron
echo "declare -x OS_PASSWORD=\"${OS_PASSWORD}\"" >> /env.cron
if [ "$BACKUP_INFLUXDB_FULL" ] ; then
  echo "declare -x BACKUP_INFLUXDB_FULL=\"${BACKUP_INFLUXDB_FULL}\"" >> /env.cron
elif [ "$BACKUP_PGSQL_FULL" ] ; then
  echo "declare -x BACKUP_PGSQL_FULL=\"${BACKUP_PGSQL_FULL}\"" >> /env.cron
  echo "declare -x BACKUP_PGSQL_INCR=\"${BACKUP_PGSQL_INCR}\"" >> /env.cron
else
  echo "declare -x BACKUP_MYSQL_FULL=\"${BACKUP_MYSQL_FULL}\"" >> /env.cron
  echo "declare -x BACKUP_MYSQL_INCR=\"${BACKUP_MYSQL_INCR}\"" >> /env.cron
  echo "declare -x MYSQL_ROOT_PASSWORD=\"${MYSQL_ROOT_PASSWORD}\"" >> /env.cron
fi
chmod 600 /env.cron
cron -f
