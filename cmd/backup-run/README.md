# Backup Setup

To setup the backup for helm-charts you only need this:
Where every you need a backup for postgres, you have to add the following code to your deployment:

## helm-charts

In the relevant chart, the following changes are to be made through to setting up:

To finde the backup image_version `grep -r '  image_version' .`

### charts/postgres/values.yaml - default settings - if not already used a requirment depencie -

This config is the default value with a disabled backup.

```
postgres:
  ...
  backup:
    enabled: false
    interval_full: 1 days
    interval_incr: 1 hours
    image_version: v0.5.6
    os_password: DEFINED-IN-REGION-SECRETS
    os_username: db_backup
    os_user_domain: Default
    os_project_name: master
    os_project_domain: ccadmin
```

##### charts/postgres/templates/deployment.yaml - default settings

This needs to be added:

```
spec:
...
    spec:
      containers:
      ... postgres and other containers ...
{{- if .Values.backup.enabled }}
      - image: {{.Values.backup.repository}}:{{.Values.backup.image_version}}
        name: backup
        env:
          - name: MY_POD_NAME
            value: {{ template "fullname" . }}
          - name: MY_POD_NAMESPACE
            value: {{.Release.Namespace}}
          - name: OS_AUTH_URL
            value: {{.Values.backup.os_auth_url}}
          - name: OS_AUTH_VERSION
            value: "3"
          - name: OS_USERNAME
            value: {{.Values.backup.os_username}}
          - name: OS_USER_DOMAIN_NAME
            value: {{.Values.backup.os_user_domain}}
          - name: OS_PROJECT_NAME
            value: {{.Values.backup.os_project_name}}
          - name: OS_PROJECT_DOMAIN_NAME
            value: {{.Values.backup.os_project_domain}}
          - name: OS_REGION_NAME
            value: {{.Values.backup.os_region_name}}
          - name: OS_PASSWORD
            value: {{.Values.backup.os_password | quote}}
          - name: BACKUP_PGSQL_FULL
            value: {{.Values.backup.interval_full | quote}}
{{- end }}
...
```
