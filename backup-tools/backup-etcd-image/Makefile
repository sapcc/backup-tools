SHELL       := /bin/sh
REGISTRY    := $(if $(REGISTRY),$(REGISTRY)/,)
IMAGE       ?= sapcc/backup-etcd
VERSION     ?= v0.0.1
DOCKERFILE  ?= Dockerfile.backup-etcd-image

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
	docker build $(BUILD_ARGS) -f ../../$(DOCKERFILE) -t $(REGISTRY)$(IMAGE):$(VERSION) --rm ../..
	docker system prune -f

.PHONY: push
push:
	docker push $(REGISTRY)$(IMAGE):$(VERSION)
