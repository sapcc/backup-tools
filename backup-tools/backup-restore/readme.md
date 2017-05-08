# backup-restore


### mariadb, mysql and postgres backups

To restore a mysql/mariadb or postgres backup please run in your `backup` container of your pod this command:

`backup-restore`

follow the instructions there. At any error the process will give back a fatal error and stop the process.

### influxdb backups

You need to set the **"backup_restore"** switch and re-deploy the influxdb pod.

To restore an influxdb backup please run first in your `backup` container of your pod this command:

`backup-restore ic`

copy the output string, you will need it in the  next setp. Now you need to `attach` the influxdb pod. There you need to follow the instructions and here you will need the output string from the command before. Past it there and **_please make sure that the string have no line breaks or additional spaces_**!


## Build Go binary and update configs

If you need to change the code please build it for `Linux/amd64` and upload the binary to the "release". To make a release tag the current git status with `restore-v....`. Now you can add there the binary.

Now you need to update your helm deployment and add change the download url.