# backup-restore


### mariadb, mysql and postgres backups

To restore a mysql/mariadb or postgres backup please run in your `backup` container of your pod this command:

`backup-restore`

follow the instructions there. At any error the process will give back a fatal error and stop the process.

### Cross-Region Backup-Restore
Run this in the POD container `backup`.

To minifiy the config to run a cross-region backup you can create a config with the following command from the same pod in an other region.

To create a config for restore backups for `eu-de-1` from region `eu-nl-1` run the following on the same POD in region `eu-nl-1`:

`BACKUP_REGION_NAME="eu-de-1" backup-restore cc`

copy the output string, you will need it in the  next setp. Now you need to `attach` the pod in `eu-de-1`. There you need to follow the instructions and here you will need the output string from the command before. In the process you need to enter `crossregion` where you can paste the config string. Past it there and **_please make sure that the string have no line breaks or additional spaces_**!

Now you get the list from swift in `eu-nl-1` where the backup from `eu-de-1` is replicated. Select your backup that you like to restore.

Thats all.

If you don't like to use the `backup-restore cc` command in an other region from where you like to restore your backup, then you can also start the process with this command:

`CONTAINER_PREFIX="eu-de-1/c5252118/pgsql" OS_AUTH_URL="url to the auth endpoint" OS_USERNAME="backup" OS_PASSWORD="password for backupuser" OS_USER_DOMAIN_NAME="userdomainname" OS_PROJECT_NAME="projectname" OS_PROJECT_DOMAIN_NAME="projectdomainname" OS_REGION_NAME="region i.e. eu-de-1" backup-restore`

This ENV variables need to be give with the `backup-restore` command
```
CONTAINER_PREFIX="REGION_OF_BACKUP_TO_RESTORE/POD_NAMESPACE/POD_NAME"
OS_AUTH_URL="url to the auth endpoint from where we download our backup-replica"
OS_USERNAME="backupuser of the swift from where we download our backup-replica"
OS_USER_DOMAIN_NAME="userdomainname of the swift from where we download our backup-replica"
OS_PROJECT_NAME="projectname of the swift from where we download our backup-replica"
OS_PROJECT_DOMAIN_NAME="projectdomainname of the swift from where we download our backup-replica"
OS_REGION_NAME="region i.e. eu-de-1 of the swift from where we download our backup-replica"
OS_PASSWORD="password for backupuser of the swift from where we download our backup-replica"
```

### Cross-Region Manual-Restore

You will be able to restore manual backups.

Please create the directory `/newbackup/` in the root path of container `backup` in your POD. Example:

`monsoonctl exec pgsql --namespace c5252118 -i -c backup -- /bin/bash -c 'mkdir /newbackup/'`

After that, you can transfer your backup files to it. Example:

`cat backup.tar.gz | monsoonctl exec pgsql --namespace c5252118 -i -c backup -- /bin/bash -c 'cat >/newbackup/backup.tar.gz'`
`cat backup.zip | monsoonctl exec pgsql --namespace c5252118 -i -c backup -- /bin/bash -c 'cat >/newbackup/backup.zip'`
`cat backup.sql | monsoonctl exec pgsql --namespace c5252118 -i -c backup -- /bin/bash -c 'cat >/newbackup/backup.sql'`

When you have tranfered all your needed backup files, you can run `backup-restore` and enter `manual` to restore your backup from `/newbackup/`


## Build Go binary and update configs

If you need to change the code please build it for `Linux/amd64` and upload the binary to the "release". To make a release tag of the current git status with `restore-v....` i.e. `restore-v0.2.0`. Now you can add there the binary.

Now you need to update your helm deployment and change the download url.

