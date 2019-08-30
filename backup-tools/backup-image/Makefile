SHELL       := /bin/sh
REGISTRY    := $(if $(REGISTRY),$(REGISTRY)/,)
IMAGE       ?= sapcc/backup-tools
VERSION     ?= v0.6.1
DOCKERFILE  ?= Dockerfile.backup-image

### Proxy Foo
ifneq ($(http_proxy),)
BUILD_ARGS += --build-arg http_proxy=$(http_proxy) --build-arg https_proxy=$(https_proxy) --build-arg no_proxy=$(no_proxy)
endif
ifneq ($(NO_CACHE),)
BUILD_ARGS += --no-cache
endif

### Docker Targets

all: build

.PHONY: build
build:
	docker build $(BUILD_ARGS) -f ../../$(DOCKERFILE) -t $(REGISTRY)$(IMAGE):$(VERSION) --rm ../..

.PHONY: push
push:
	docker push $(REGISTRY)$(IMAGE):$(VERSION)
