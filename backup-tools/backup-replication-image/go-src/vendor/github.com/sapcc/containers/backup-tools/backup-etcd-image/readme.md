# HELM image
Install this via HELM in namespace: `backup`
<br>
helm-chart: `backup-etcd`

Build a Debug version with Busybox

`IMAGE=sapcc/backup-etcd VERSION=v0.0.17 DOCKERFILE=Dockerfile.debug make build`