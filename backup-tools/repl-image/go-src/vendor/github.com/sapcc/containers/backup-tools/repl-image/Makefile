SHELL       := /bin/sh
REGISTRY    := $(if $(REGISTRY),$(REGISTRY)/,)
IMAGE       ?= sapcc/replication-tools
VERSION     ?= v0.1.2

### Proxy Foo
ifneq ($(http_proxy),)
BUILD_ARGS += --build-arg http_proxy=$(http_proxy) --build-arg https_proxy=$(https_proxy) --build-arg no_proxy=$(no_proxy)
endif
ifneq ($(NO_CACHE),)
BUILD_ARGS += --no-cache
endif

### Docker Targets

.PHONY: build
build:
	mkdir -p go-src/files && tar -cvf go-src/files/go-src.tar --exclude ".git" --exclude ".cache/" --exclude "*.tar" -C ~/go/ .
	docker build $(BUILD_ARGS) --squash -t go-build:latest --rm go-src/
	test -f go-src/files/go-src.tar && rm go-src/files/go-src.tar
	docker run --rm go-build tar -cvf - /replication-run /backup/env/config.yml > files/replication-run.tar
	docker run --rm go-build tar -cvf - /etc/ssl/certs/ca-certificates.crt > files/backup-certs.tar
	docker rmi go-build
	docker build $(BUILD_ARGS) --squash -t $(REGISTRY)$(IMAGE):$(VERSION) --rm .
	docker system prune -f

.PHONY: push
push:
	docker push $(REGISTRY)/$(IMAGE):$(VERSION)
