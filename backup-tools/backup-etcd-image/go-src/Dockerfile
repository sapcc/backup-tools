FROM golang

COPY backup-etcd.go /go/src/backup-etcd/

RUN set -uex; \
    cd /go/src/backup-etcd; \
    export GOPATH=/go; \
    go get; \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /go/bin/backup-etcd; \
    mv /go/bin/backup-etcd /
