# SPDX-FileCopyrightText: 2025 SAP SE or an SAP affiliate company
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.24.5-alpine3.22 AS builder

RUN apk add --no-cache --no-progress ca-certificates gcc git make musl-dev

COPY . /src
ARG BININFO_BUILD_DATE BININFO_COMMIT_HASH BININFO_VERSION # provided to 'make install'
RUN make -C /src install PREFIX=/pkg GOTOOLCHAIN=local GO_BUILDFLAGS='-mod vendor'

################################################################################

FROM alpine:3.22

RUN addgroup -g 4200 appgroup \
  && adduser -h /home/appuser -s /sbin/nologin -G appgroup -D -u 4200 appuser

# upgrade all installed packages to fix potential CVEs in advance
# also remove apk package manager to hopefully remove dependency on OpenSSL 🤞
RUN apk upgrade --no-cache --no-progress \
  && apk add --no-cache --no-progress postgresql15-client postgresql16-client curl jq \
  && apk del --no-cache --no-progress apk-tools alpine-keys alpine-release libc-utils

COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs/
COPY --from=builder /etc/ssl/cert.pem /etc/ssl/cert.pem
COPY --from=builder /pkg/ /usr/
# make sure all binaries can be executed
RUN set -x \
  && backup-server --version 2>/dev/null

ARG BININFO_BUILD_DATE BININFO_COMMIT_HASH BININFO_VERSION
LABEL source_repository="https://github.com/sapcc/containers" \
  org.opencontainers.image.url="https://github.com/sapcc/containers" \
  org.opencontainers.image.created=${BININFO_BUILD_DATE} \
  org.opencontainers.image.revision=${BININFO_COMMIT_HASH} \
  org.opencontainers.image.version=${BININFO_VERSION}

ENV ENV=/usr/bin/motd.sh
COPY --from=postgres:17-alpine /usr/local/bin/pg_dump /usr/libexec/postgresql17/pg_dump
RUN /usr/libexec/postgresql17/pg_dump --version >/dev/null
USER 4200:4200
WORKDIR /home/appuser
ENTRYPOINT [ "/usr/bin/backup-server" ]
