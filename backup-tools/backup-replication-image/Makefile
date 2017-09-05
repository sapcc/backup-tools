SHELL       := /bin/sh
REGISTRY    := $(if $(REGISTRY),$(REGISTRY)/,)
IMAGE       ?= sapcc/replication-tools
VERSION     ?= v0.1.2
DOCKERFILE  ?= Dockerfile.backup-replication-image

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
	docker build $(BUILD_ARGS) --squash -f ../../$(DOCKERFILE) -t $(REGISTRY)$(IMAGE):$(VERSION) --rm ../..
	docker system prune -f

.PHONY: push
push:
	docker push $(REGISTRY)$(IMAGE):$(VERSION)
