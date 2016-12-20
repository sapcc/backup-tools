FROM hub.global.cloud.sap/monsoon/ubuntu:test16.04
MAINTAINER "Josef Fröhle <josef.froehle@sap.com>, Norbert Tretkowski <norbert.tretkowski@sap.com>"

ADD scripts/db-backup.sh /usr/local/sbin/db-backup.sh
ADD additional_data/crontab /etc/cron.d/db-backup

RUN mkdir /backup \
	&& sed -i s/^deb-src/\#\ deb-src/g /etc/apt/sources.list \
	&& sed -i s/archive\.ubuntu\.com/de\.archive\.ubuntu\.com/g /etc/apt/sources.list \
	&& echo "APT::Install-Suggests "0";" > /etc/apt/apt.conf.d/99local \
	&& echo "APT::Install-Recommends "0";" >> /etc/apt/apt.conf.d/99local \
	&& chmod a+x /usr/local/sbin/db-backup.sh \
	&& apt-get update && apt-get upgrade -y \
	&& apt-get install -y --no-install-recommends wget lsb-release ca-certificates \
	&& wget https://repo.percona.com/apt/percona-release_0.1-4.$(lsb_release -sc)_all.deb \
	&& dpkg -i percona-release_0.1-4.$(lsb_release -sc)_all.deb \
	&& wget -ct0 http://ftp.de.debian.org/debian/pool/main/p/python-msrest/python3-msrest_0.4.4-1_all.deb \
	&& wget -ct0 http://ftp.de.debian.org/debian/pool/main/p/python-msrestazure/python3-msrestazure_0.4.4-1_all.deb \
	&& wget -ct0 http://ftp.de.debian.org/debian/pool/main/p/python-azure/python3-azure_2.0.0~rc6+dfsg-2_all.deb \
	&& wget -ct0 http://ftp.de.debian.org/debian/pool/main/p/python-gevent/python3-gevent_1.1.2-1_amd64.deb \
	&& wget -ct0 http://ftp.de.debian.org/debian/pool/main/p/python-boto/python3-boto_2.43.0-1_all.deb \
	&& wget -ct0 http://ftp.de.debian.org/debian/pool/main/p/python-requests-oauthlib/python3-requests-oauthlib_0.7.0-0.1_all.deb \
	&& apt-get --no-install-recommends -y install python3-gevent python3-boto python3-swiftclient python3-swiftclient lzop pv python3-certifi python3-isodate python3-requests-oauthlib \
	&& dpkg -i python3-requests-oauthlib_0.7.0-0.1_all.deb \
	&& dpkg -i python3-msrest_0.4.4-1_all.deb \
	&& dpkg -i python3-msrestazure_0.4.4-1_all.deb \
	&& dpkg -i python3-azure_2.0.0~rc6+dfsg-2_all.deb \
	&& dpkg -i python3-gevent_1.1.2-1_amd64.deb \
	&& dpkg -i python3-boto_2.43.0-1_all.deb \
	&& apt-get -f install \
	&& echo "deb http://apt.postgresql.org/pub/repos/apt/ jessie-pgdg main 9.6" > /etc/apt/sources.list.d/pgdg.list \
	&& wget --quiet -O - https://www.postgresql.org/media/keys/ACCC4CF8.asc | apt-key add - \
	&& apt-get update && apt-get dist-upgrade -y \
	&& apt-get install -y --no-install-recommends mariadb-client postgresql-client python-swiftclient cron barman percona-xtrabackup-24 \
	&& apt-get install -y --no-install-recommends less vim iproute2 man-db mc \
	&& rm *.deb \
	&& rm -rf /var/lib/apt/lists/* \
	&& chmod 0644 /etc/cron.d/db-backup \
	&& touch /var/log/cron.log

VOLUME ["/backup"]

CMD ["cron", "-f"]
