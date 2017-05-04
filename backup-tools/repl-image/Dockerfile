FROM hub.global.cloud.sap/monsoon/ubuntu:test16.04

MAINTAINER "Josef Fröhle <josef.froehle@sap.com>, Norbert Tretkowski <norbert.tretkowski@sap.com>"

ADD files/backup-replication.sh /usr/local/sbin/backup-replication.sh

ADD files/replication-run.tar /

RUN mkdir /backup \
	&& sed -i s/^deb-src/\#\ deb-src/g /etc/apt/sources.list \
	&& sed -i s/archive\.ubuntu\.com/de\.archive\.ubuntu\.com/g /etc/apt/sources.list \
	&& chmod a+x /usr/local/sbin/backup-replication.sh \
	&& chmod a+x /usr/local/sbin/replication-run \
	&& echo "APT::Install-Suggests "0";" > /etc/apt/apt.conf.d/99local \
	&& echo "APT::Install-Recommends "0";" >> /etc/apt/apt.conf.d/99local \
	&& apt-get update && apt-get upgrade -y \
	&& apt-get install -y --no-install-recommends curl lsb-release ca-certificates \
	&& apt-get install -y --no-install-recommends python-swiftclient \
	&& apt-get install -y --no-install-recommends less vim iproute2 man-db mc \
	&& rm -f /var/log/*.log /var/log/apt/* \
	&& rm -rf /var/lib/apt/lists/*

VOLUME ["/backup"]

CMD ["/usr/local/sbin/replication-run"]
