FROM scratch

MAINTAINER "Josef Fröhle <josef.froehle@sap.com>, Norbert Tretkowski <norbert.tretkowski@sap.com>"

ADD files/replication-run.tar /
ADD files/backup-certs.tar /

VOLUME ["/backup"]

ENTRYPOINT ["/replication-run"]
