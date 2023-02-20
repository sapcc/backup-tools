FROM golang:1.20 as gobuilder

COPY . /src
RUN cd /src && CGO_ENABLED=0 go build -ldflags "-s -w" -o /pkg/bin/backup-restore ./cmd/backup-restore
RUN cd /src && CGO_ENABLED=0 go build -ldflags "-s -w" -o /pkg/bin/backup-tools   .

# TODO: Once we've updated all Postgres to beyond 12, move to an Alpine image with psql/pg_dump from the stock Alpine packages.
FROM ubuntu:22.04
MAINTAINER "Josef Fröhle <josef.froehle@sap.com>, Norbert Tretkowski <norbert.tretkowski@sap.com>"
LABEL source_repository="https://github.com/sapcc/containers"

ENV RESTOREVER=0.1.0
ENV TZ=Etc/UTC
ARG DEBIAN_FRONTEND=noninteractive
ARG POSTGRES_VERSION=12

RUN mkdir /backup \
	&& sed -i s/^deb-src/\#\ deb-src/g /etc/apt/sources.list \
	&& sed -i s/archive\.ubuntu\.com/de\.archive\.ubuntu\.com/g /etc/apt/sources.list \
	&& echo "APT::Install-Suggests "0";" > /etc/apt/apt.conf.d/99local \
	&& echo "APT::Install-Recommends "0";" >> /etc/apt/apt.conf.d/99local \
	&& apt-get update && apt-get upgrade -y \
	&& apt-get install -y --no-install-recommends wget lsb-release ca-certificates gnupg2 \
	&& echo "deb http://apt.postgresql.org/pub/repos/apt/ jammy-pgdg main $POSTGRES_VERSION" > /etc/apt/sources.list.d/postgresql.list \
	&& wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
	&& apt-get update && apt-get dist-upgrade -y \
	&& apt-get install -y --no-install-recommends postgresql-client-$POSTGRES_VERSION python3-swiftclient \
	&& apt-get install -y --no-install-recommends less vim iproute2 \
	&& rm -f /var/log/*.log /var/log/apt/* \
	&& rm -rf /var/lib/apt/lists/* \
	&& ln -sf /proc/1/fd/1 /var/log/backup.log \
	&& test -x /usr/bin/swift \
	&& test -x /usr/bin/psql \
	&& test -x /usr/bin/pg_dump

COPY --from=gobuilder /pkg/bin/ /usr/local/sbin/

VOLUME ["/backup"]
CMD ["/usr/local/sbin/backup-tools", "create"]
