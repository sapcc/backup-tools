FROM golang:1.20-alpine3.17 as gobuilder

COPY . /src
RUN cd /src && go build -ldflags "-s -w" -o /pkg/bin/backup-restore ./cmd/backup-restore
RUN cd /src && go build -ldflags "-s -w" -o /pkg/bin/backup-tools   .

FROM alpine:3.17
LABEL source_repository="https://github.com/sapcc/containers"
ARG POSTGRES_VERSION=12

RUN mkdir /backup \
	&& apk add --no-cache postgresql${POSTGRES_VERSION}-client ca-certificates less vim iproute2 \
	&& test -x /usr/bin/psql \
	&& test -x /usr/bin/pg_dump

COPY --from=gobuilder /pkg/bin/ /usr/local/sbin/

VOLUME ["/backup"]
CMD ["/usr/local/sbin/backup-tools", "create"]
