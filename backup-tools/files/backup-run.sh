echo "MY_POD_NAME=\"${MY_POD_NAME}\"" > /env.cron
echo "MY_POD_NAMESPACE=\"${MY_POD_NAMESPACE}\"" >> /env.cron
echo "OS_AUTH_URL=\"${OS_AUTH_URL}\"" >> /env.cron
echo "OS_AUTH_VERSION=\"${OS_AUTH_VERSION}\"" >> /env.cron
echo "OS_IDENTITY_API_VERSION=\"${OS_IDENTITY_API_VERSION}\"" >> /env.cron
echo "OS_USERNAME=\"${OS_USERNAME}\"" >> /env.cron
echo "OS_USER_DOMAIN_NAME=\"${OS_USER_DOMAIN_NAME}\"" >> /env.cron
echo "OS_PROJECT_NAME=\"${OS_PROJECT_NAME}\"" >> /env.cron
echo "OS_PROJECT_DOMAIN_NAME=\"${OS_PROJECT_DOMAIN_NAME}\"" >> /env.cron
echo "OS_REGION_NAME=\"${OS_REGION_NAME}\"" >> /env.cron
echo "OS_PASSWORD=\"${OS_PASSWORD}\"" >> /env.cron
echo "BACKUP_PGSQL_FULL=\"${BACKUP_PGSQL_FULL}\"" >> /env.cron
echo "BACKUP_PGSQL_INCR=\"${BACKUP_PGSQL_INCR}\"" >> /env.cron

cron -f
