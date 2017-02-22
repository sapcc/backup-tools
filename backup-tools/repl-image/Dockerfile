FROM hub.global.cloud.sap/monsoon/ubuntu:test16.04

MAINTAINER "Josef Fröhle <josef.froehle@sap.com>, Norbert Tretkowski <norbert.tretkowski@sap.com>"

ADD files/replication-run.sh /usr/local/sbin/replication-run.sh

RUN mkdir /backup \
	&& sed -i s/^deb-src/\#\ deb-src/g /etc/apt/sources.list \
	&& sed -i s/archive\.ubuntu\.com/de\.archive\.ubuntu\.com/g /etc/apt/sources.list \
	&& echo "APT::Install-Suggests "0";" > /etc/apt/apt.conf.d/99local \
	&& echo "APT::Install-Recommends "0";" >> /etc/apt/apt.conf.d/99local \
	&& apt-get update && apt-get upgrade -y \
	&& apt-get install -y --no-install-recommends wget lsb-release ca-certificates \
	&& apt-get install -y --no-install-recommends python-swiftclient cron \
	&& apt-get install -y --no-install-recommends less vim iproute2 man-db mc \
	&& rm -rf /var/lib/apt/lists/* \
	&& ln -sf /proc/1/fd/1 /var/log/replication.log \
	&& touch /var/log/cron.log

VOLUME ["/backup"]

CMD ["/usr/local/sbin/replication-run.sh"]
