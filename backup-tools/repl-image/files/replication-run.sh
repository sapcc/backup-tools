#!/bin/bash
echo "declare -x MY_POD_NAME=\"${MY_POD_NAME}\"" > /env.cron
echo "declare -x MY_POD_NAMESPACE=\"${MY_POD_NAMESPACE}\"" >> /env.cron
echo "declare -x OS_AUTH_URL=\"${OS_AUTH_URL}\"" >> /env.cron
echo "declare -x OS_AUTH_VERSION=\"${OS_AUTH_VERSION}\"" >> /env.cron
echo "declare -x OS_IDENTITY_API_VERSION=\"${OS_IDENTITY_API_VERSION}\"" >> /env.cron
echo "declare -x OS_USERNAME=\"${OS_USERNAME}\"" >> /env.cron
echo "declare -x OS_USER_DOMAIN_NAME=\"${OS_USER_DOMAIN_NAME}\"" >> /env.cron
echo "declare -x OS_PROJECT_NAME=\"${OS_PROJECT_NAME}\"" >> /env.cron
echo "declare -x OS_PROJECT_DOMAIN_NAME=\"${OS_PROJECT_DOMAIN_NAME}\"" >> /env.cron
echo "declare -x OS_REGION_NAME=\"${OS_REGION_NAME}\"" >> /env.cron
echo "declare -x OS_PASSWORD=\"${OS_PASSWORD}\"" >> /env.cron
chmod 600 /env.cron
cron -f
