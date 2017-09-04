FROM golang

COPY . /go/src/github.com/sapcc/containers/backup-tools/backup-restore

RUN set -uex; \
    cd /go/src/github.com/sapcc/containers/backup-tools/backup-restore; \
    export GOPATH=/go; \
    go get; \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /go/bin/backup-restore; \
    mv /go/bin/backup-restore /usr/local/sbin
