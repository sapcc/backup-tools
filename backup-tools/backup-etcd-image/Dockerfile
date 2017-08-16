FROM quay.io/coreos/etcd:v2.3.8

MAINTAINER "Josef Fröhle <josef.froehle@sap.com>, Norbert Tretkowski <norbert.tretkowski@sap.com>"

ADD files/backup-etcd.tar /

VOLUME ["/backup"]

ENTRYPOINT ["/backup-etcd"]
