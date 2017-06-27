FROM hub.global.cloud.sap/monsoon/ubuntu:test16.04

MAINTAINER "Josef Fröhle <josef.froehle@sap.com>, Norbert Tretkowski <norbert.tretkowski@sap.com>"

ENV RESTOREVER=0.1.0

ADD files/db-backup.sh /usr/local/sbin/db-backup.sh
#ADD files/backup-run.sh /usr/local/sbin/backup-run.sh

ADD files/backup-run.tar /

RUN mkdir /backup \
	&& sed -i s/^deb-src/\#\ deb-src/g /etc/apt/sources.list \
	&& sed -i s/archive\.ubuntu\.com/de\.archive\.ubuntu\.com/g /etc/apt/sources.list \
	&& echo "APT::Install-Suggests "0";" > /etc/apt/apt.conf.d/99local \
	&& echo "APT::Install-Recommends "0";" >> /etc/apt/apt.conf.d/99local \
	&& chmod a+x /usr/local/sbin/db-backup.sh \
	&& chmod a+x /usr/local/sbin/backup-run \
	&& apt-get update && apt-get upgrade -y \
	&& apt-get install -y --no-install-recommends wget lsb-release ca-certificates \
	&& echo "deb http://apt.postgresql.org/pub/repos/apt/ jessie-pgdg main 9.6" > /etc/apt/sources.list.d/postgresql.list \
	&& wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
	&& apt-get update && apt-get dist-upgrade -y \
	&& apt-get install -y --no-install-recommends mariadb-client postgresql-client python-swiftclient \
	&& apt-get install -y --no-install-recommends less vim iproute2 man-db mc \
	&& wget -O /usr/local/bin/backup-restore https://github.com/sapcc/containers/releases/download/restore-v${RESTOREVER}/backup-restore.linux_amd64 \
	&& chmod a+x /usr/local/bin/backup-restore \
	&& rm *.deb \
	&& rm -f /var/log/*.log /var/log/apt/* \
	&& rm -rf /var/lib/apt/lists/* \
	&& ln -sf /proc/1/fd/1 /var/log/backup.log \
	&& test -x /usr/bin/swift \
	&& test -x /usr/bin/mysql \
	&& test -x /usr/bin/mysqldump \
	&& test -x /usr/bin/psql \
	&& test -x /usr/bin/pg_dump

VOLUME ["/backup"]

CMD ["/usr/local/sbin/backup-run"]
