FROM golang as gobuilder

COPY go-src/ /go/src/replication-run/

RUN set -uex; \
    cd /go/src/replication-run; \
    export GOPATH=/go; \
    go get ... ; \
    CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-s -w' -installsuffix cgo -o /go/bin/replication-run; \
    mkdir -p /backup/env 

FROM scratch
MAINTAINER "Josef Fröhle <josef.froehle@sap.com>, Norbert Tretkowski <norbert.tretkowski@sap.com>"

COPY --from=gobuilder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=gobuilder /go/bin/replication-run /replication-run
COPY --from=gobuilder /backup/env /backup/env

VOLUME ["/backup"]

ENTRYPOINT ["/replication-run"]
