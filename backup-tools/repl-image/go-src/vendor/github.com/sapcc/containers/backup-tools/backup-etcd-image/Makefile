SHELL       := /bin/sh
REGISTRY    := $(if $(REGISTRY),$(REGISTRY)/,)
IMAGE       ?= sapcc/backup-etcd
VERSION     ?= v0.0.1
DOCKERFILE  ?= Dockerfile

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
	tar -cvf go-src/files/go-src.tar --exclude ".git" --exclude ".cache/" --exclude "*.tar" -C ~/go/ .
	docker build $(BUILD_ARGS) --squash -t go-build:latest --rm ./go-src
	test -f go-src/files/go-src.tar && rm go-src/files/go-src.tar
	docker run --rm go-build tar -cvf - /backup-etcd > files/backup-etcd.tar
	docker run --rm go-build tar -cvf - /etc/ssl/certs > files/backup-certs.tar
	docker rmi go-build:latest
	docker build $(BUILD_ARGS) --squash -f $(DOCKERFILE) -t $(REGISTRY)$(IMAGE):$(VERSION) --rm .
	test -f files/backup-etcd.tar && rm files/backup-etcd.tar
	docker system prune -f

.PHONY: push
push:
	docker push $(REGISTRY)$(IMAGE):$(VERSION)
