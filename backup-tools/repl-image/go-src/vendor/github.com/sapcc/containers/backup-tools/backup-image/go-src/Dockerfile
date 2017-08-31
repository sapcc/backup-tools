FROM golang

COPY backup-run.go /go/src/backup-run/

RUN set -uex; \
    cd /go/src/backup-run; \
    export GOPATH=/go; \
    go get; \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /go/bin/backup-run; \
    mv /go/bin/backup-run /usr/local/sbin
