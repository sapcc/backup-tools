FROM golang

ADD files/go-src.tar /go/

RUN set -uex; \
    cd /go/src/github.com/sapcc/containers/backup-tools/backup-etcd-image/go-src; \
    export GOPATH=/go; \
    go get ./...; \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /go/bin/backup-etcd; \
    mv /go/bin/backup-etcd /
