SHELL       := /bin/sh
REGISTRY    ?=
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
	docker build -t go-build:latest --rm go-src/
	docker run go-build tar -cvf - /usr/local/sbin/replication-run > files/replication-run.tar
	docker build $(BUILD_ARGS) -t $(REGISTRY)/$(IMAGE):$(VERSION) --rm .

.PHONY: push
push:
	docker push $(REGISTRY)/$(IMAGE):$(VERSION)
