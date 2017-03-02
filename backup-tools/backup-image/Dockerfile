FROM hub.global.cloud.sap/monsoon/ubuntu:test16.04

MAINTAINER "Josef Fröhle <josef.froehle@sap.com>, Norbert Tretkowski <norbert.tretkowski@sap.com>"

ENV INFLUXVER=1.2.0

ADD files/crontab /etc/cron.d/db-backup
ADD files/db-backup.sh /usr/local/sbin/db-backup.sh
ADD files/backup-run.sh /usr/local/sbin/backup-run.sh

RUN mkdir /backup \
	&& sed -i s/^deb-src/\#\ deb-src/g /etc/apt/sources.list \
	&& sed -i s/archive\.ubuntu\.com/de\.archive\.ubuntu\.com/g /etc/apt/sources.list \
	&& echo "APT::Install-Suggests "0";" > /etc/apt/apt.conf.d/99local \
	&& echo "APT::Install-Recommends "0";" >> /etc/apt/apt.conf.d/99local \
	&& chmod a+x /usr/local/sbin/db-backup.sh \
	&& chmod a+x /usr/local/sbin/backup-run.sh \
	&& echo "deb http://tretkowski.de/sap/ backup-tools/" >> /etc/apt/sources.list.d/sap-repo.list \
	&& apt-get update && apt-get upgrade -y \
	&& apt-get install -y --no-install-recommends wget lsb-release ca-certificates \
	&& wget https://repo.percona.com/apt/percona-release_0.1-4.$(lsb_release -sc)_all.deb \
	&& dpkg -i percona-release_0.1-4.$(lsb_release -sc)_all.deb \
	&& apt-get install -y --no-install-recommends --allow-unauthenticated wal-e \
	&& echo "deb http://apt.postgresql.org/pub/repos/apt/ jessie-pgdg main 9.6" > /etc/apt/sources.list.d/postgresql.list \
	&& wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
	&& apt-get update && apt-get dist-upgrade -y \
	&& apt-get install -y --no-install-recommends mariadb-client python-swiftclient cron barman percona-xtrabackup-24 \
	&& apt-get install -y --no-install-recommends less vim iproute2 man-db mc \
	&& dpkg --purge postgresql-client-9.5 \
	&& wget "https://dl.influxdata.com/influxdb/releases/influxdb_${INFLUXVER}_amd64.deb" \
	&& dpkg -i influxdb_${INFLUXVER}_amd64.deb \
	&& rm *.deb \
	&& rm -f /var/log/*.log /var/log/apt/* \
	&& rm -rf /var/lib/apt/lists/* \
	&& chmod 0644 /etc/cron.d/db-backup \
	&& ln -sf /proc/1/fd/1 /var/log/backup.log \
	&& touch /var/log/cron.log

VOLUME ["/backup"]

CMD ["/usr/local/sbin/backup-run.sh"]
