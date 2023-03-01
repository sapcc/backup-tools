FROM golang:1.20-alpine3.17 as gobuilder

COPY . /src
RUN cd /src && go build -ldflags "-s -w" -mod vendor -o /pkg/bin/backup-restore ./cmd/backup-restore
RUN cd /src && go build -ldflags "-s -w" -mod vendor -o /pkg/bin/backup-server  .

FROM alpine:3.17
LABEL source_repository="https://github.com/sapcc/containers"
ARG POSTGRES_VERSION=12

RUN apk add --no-cache --no-progress postgresql${POSTGRES_VERSION}-client ca-certificates curl jq less vim iproute2

COPY --from=gobuilder /pkg/bin/ /usr/local/sbin/
COPY ./backup-tools.sh          /usr/local/sbin/backup-tools

CMD ["/usr/local/sbin/backup-server"]
