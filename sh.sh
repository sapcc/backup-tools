#!/bin/sh

###
### This script is installed as /usr/bin/sh in the Docker container.
### It is not used when calling backup-tools.sh, but when the user enters
### `kubectl exec $POD -- sh` or such, this entrypoint will run and show
### the backup-tools help as a banner (similar to MOTD).
###

backup-tools help
exec /bin/sh "$@"
