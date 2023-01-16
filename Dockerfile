FROM golang as gobuilder

COPY . /src
RUN cd /src && CGO_ENABLED=0 go build -ldflags "-s -w" -o /pkg/bin/backup-run     ./cmd/backup-run
RUN cd /src && CGO_ENABLED=0 go build -ldflags "-s -w" -o /pkg/bin/backup-restore ./cmd/backup-restore

FROM ubuntu:22.04
MAINTAINER "Josef Fröhle <josef.froehle@sap.com>, Norbert Tretkowski <norbert.tretkowski@sap.com>"
LABEL source_repository="https://github.com/sapcc/containers"

ENV RESTOREVER=0.1.0
ENV TZ=Etc/UTC
ARG DEBIAN_FRONTEND=noninteractive

RUN mkdir /backup \
	&& sed -i s/^deb-src/\#\ deb-src/g /etc/apt/sources.list \
	&& sed -i s/archive\.ubuntu\.com/de\.archive\.ubuntu\.com/g /etc/apt/sources.list \
	&& echo "APT::Install-Suggests "0";" > /etc/apt/apt.conf.d/99local \
	&& echo "APT::Install-Recommends "0";" >> /etc/apt/apt.conf.d/99local \
	&& apt-get update && apt-get upgrade -y \
	&& apt-get install -y --no-install-recommends wget lsb-release ca-certificates gnupg2 \
	&& echo "deb http://apt.postgresql.org/pub/repos/apt/ jammy-pgdg main 12" > /etc/apt/sources.list.d/postgresql.list \
	&& wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
	&& apt-get update && apt-get dist-upgrade -y \
	&& apt-get install -y --no-install-recommends mariadb-client postgresql-client python3-swiftclient \
	&& apt-get install -y --no-install-recommends less vim iproute2 man-db mc \
	&& rm -f /var/log/*.log /var/log/apt/* \
	&& rm -rf /var/lib/apt/lists/* \
	&& ln -sf /proc/1/fd/1 /var/log/backup.log \
	&& test -x /usr/bin/swift \
	&& test -x /usr/bin/mysql \
	&& test -x /usr/bin/mysqldump \
	&& test -x /usr/bin/psql \
	&& test -x /usr/bin/pg_dump

COPY --from=gobuilder /pkg/bin/ /usr/local/sbin/
COPY ./cmd/backup-run/db-backup.sh /usr/local/sbin/db-backup.sh

VOLUME ["/backup"]
CMD ["/usr/local/sbin/backup-run"]
