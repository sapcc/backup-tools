SHELL       := /bin/sh
REGISTRY    ?=
IMAGE       ?= sapcc/backup-etcd
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
	docker build $(BUILD_ARGS) -t go-build:latest --rm ../backup-etcd
	docker run go-build tar -cvf - /backup-etcd > files/backup-etcd.tar
	docker build $(BUILD_ARGS) -t $(REGISTRY)/$(IMAGE):$(VERSION) --rm .

.PHONY: push
push:
	docker push $(REGISTRY)/$(IMAGE):$(VERSION)
