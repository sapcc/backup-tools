FROM golang

COPY replication-run.go /go/src/replication-run/

RUN set -uex; \
    cd /go/src/replication-run; \
    export GOPATH=/go; \
    go get; \
    CGO_ENABLED=0 GOOS=linux go build -ldflags "-s" -a -installsuffix cgo -o /go/bin/replication-run; \
    mv /go/bin/replication-run /usr/local/sbin
