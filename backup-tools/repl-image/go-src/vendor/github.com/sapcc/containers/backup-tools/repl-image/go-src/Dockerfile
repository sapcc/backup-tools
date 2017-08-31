FROM golang

ADD files/go-src.tar /go/
COPY . /go/src/replication-run/

RUN set -uex; \
    cd /go/src/replication-run; \
    export GOPATH=/go; \
    go get; \
    CGO_ENABLED=0 GOOS=linux go build -a -tags netgo -ldflags '-s -w' -installsuffix cgo -o /go/bin/replication-run; \
    mv /go/bin/replication-run / ; \
    mkdir -p /backup/env ; \
    echo 'from:\n\
  os_auth_url: https://identity-3.eu-de-1.cloud.sap/v3\n\
  os_auth_version: "3"\n\
  os_username: db_backup\n\
  os_user_domain: Default\n\
  os_project_name: master\n\
  os_project_domain: ccadmin\n\
  os_region_name: eu-de-1\n\
  os_password: 4Ka5hKhWYmLBs39\n\
to:\n\
- os_auth_url: https://identity-3.eu-nl-1.cloud.sap/v3\n\
  os_auth_version: "3"\n\
  os_username: db_backup\n\
  os_user_domain: Default\n\
  os_project_name: master\n\
  os_project_domain: ccadmin\n\
  os_region_name: eu-nl-1\n\
  os_password: tRR6pB5SxnpbGTeq8b2t\n\
- os_auth_url: https://identity-3.staging.cloud.sap/v3\n\
  os_auth_version: "3"\n\
  os_username: db_backup\n\
  os_user_domain: Default\n\
  os_project_name: master\n\
  os_project_domain: ccadmin\n\
  os_region_name: staging\n\
  os_password: Dw9QKthZRCUMQUf\n'\ >> /backup/env/config.yml
