################################################################################
# This file is AUTOGENERATED with <https://github.com/sapcc/go-makefile-maker> #
# Edit Makefile.maker.yaml instead.                                            #
################################################################################

MAKEFLAGS=--warn-undefined-variables
# /bin/sh is dash on Debian which does not support all features of ash/bash
# to fix that we use /bin/bash only on Debian to not break Alpine
ifneq (,$(wildcard /etc/os-release)) # check file existence
	ifneq ($(shell grep -c debian /etc/os-release),0)
		SHELL := /bin/bash
	endif
endif

default: build-all

install-scripts:
	install -d -m 0755 "$(DESTDIR)$(PREFIX)/bin"
	install -m 0755 backup-tools.sh "$(DESTDIR)$(PREFIX)/bin/backup-tools"
	install -m 0755 motd.sh         "$(DESTDIR)$(PREFIX)/bin/motd.sh"
install: install-scripts

GO_BUILDFLAGS = -mod vendor
GO_LDFLAGS =
GO_TESTENV =

# These definitions are overridable, e.g. to provide fixed version/commit values when
# no .git directory is present or to provide a fixed build date for reproducibility.
BININFO_VERSION     ?= $(shell git describe --tags --always --abbrev=7)
BININFO_COMMIT_HASH ?= $(shell git rev-parse --verify HEAD)
BININFO_BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

build-all: build/backup-server

build/backup-server: FORCE
	go build $(GO_BUILDFLAGS) -ldflags '-s -w -X github.com/sapcc/go-api-declarations/bininfo.binName=backup-server -X github.com/sapcc/go-api-declarations/bininfo.version=$(BININFO_VERSION) -X github.com/sapcc/go-api-declarations/bininfo.commit=$(BININFO_COMMIT_HASH) -X github.com/sapcc/go-api-declarations/bininfo.buildDate=$(BININFO_BUILD_DATE) $(GO_LDFLAGS)' -o build/backup-server .

DESTDIR =
ifeq ($(shell uname -s),Darwin)
	PREFIX = /usr/local
else
	PREFIX = /usr
endif

install: FORCE build/backup-server
	install -d -m 0755 "$(DESTDIR)$(PREFIX)/bin"
	install -m 0755 build/backup-server "$(DESTDIR)$(PREFIX)/bin/backup-server"

# which packages to test with "go test"
GO_TESTPKGS := $(shell go list -f '{{if or .TestGoFiles .XTestGoFiles}}{{.ImportPath}}{{end}}' ./...)
# which packages to measure coverage for
GO_COVERPKGS := $(shell go list ./...)
# to get around weird Makefile syntax restrictions, we need variables containing nothing, a space and comma
null :=
space := $(null) $(null)
comma := ,

check: FORCE static-check build/cover.html build-all
	@printf "\e[1;32m>> All checks successful.\e[0m\n"

prepare-static-check: FORCE
	@if ! hash golangci-lint 2>/dev/null; then printf "\e[1;36m>> Installing golangci-lint (this may take a while)...\e[0m\n"; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; fi
	@if ! hash go-licence-detector 2>/dev/null; then printf "\e[1;36m>> Installing go-licence-detector...\e[0m\n"; go install go.elastic.co/go-licence-detector@latest; fi
	@if ! hash addlicense 2>/dev/null; then  printf "\e[1;36m>> Installing addlicense...\e[0m\n";  go install github.com/google/addlicense@latest; fi

run-golangci-lint: FORCE prepare-static-check
	@printf "\e[1;36m>> golangci-lint\e[0m\n"
	@golangci-lint run

build/cover.out: FORCE | build
	@printf "\e[1;36m>> go test\e[0m\n"
	@env $(GO_TESTENV) go test $(GO_BUILDFLAGS) -ldflags '-s -w -X github.com/sapcc/go-api-declarations/bininfo.binName=backup-tools -X github.com/sapcc/go-api-declarations/bininfo.version=$(BININFO_VERSION) -X github.com/sapcc/go-api-declarations/bininfo.commit=$(BININFO_COMMIT_HASH) -X github.com/sapcc/go-api-declarations/bininfo.buildDate=$(BININFO_BUILD_DATE) $(GO_LDFLAGS)' -shuffle=on -p 1 -coverprofile=$@ -covermode=count -coverpkg=$(subst $(space),$(comma),$(GO_COVERPKGS)) $(GO_TESTPKGS)

build/cover.html: build/cover.out
	@printf "\e[1;36m>> go tool cover > build/cover.html\e[0m\n"
	@go tool cover -html $< -o $@

static-check: FORCE run-golangci-lint check-dependency-licenses check-license-headers

build:
	@mkdir $@

vendor: FORCE
	go mod tidy
	go mod vendor
	go mod verify

vendor-compat: FORCE
	go mod tidy -compat=$(shell awk '$$1 == "go" { print $$2 }' < go.mod)
	go mod vendor
	go mod verify

license-headers: FORCE prepare-static-check
	@printf "\e[1;36m>> addlicense\e[0m\n"
	@addlicense -c "SAP SE"  -- $(patsubst $(shell awk '$$1 == "module" {print $$2}' go.mod)%,.%/*.go,$(shell go list ./...))

check-license-headers: FORCE prepare-static-check
	@printf "\e[1;36m>> addlicense --check\e[0m\n"
	@addlicense --check  -- $(patsubst $(shell awk '$$1 == "module" {print $$2}' go.mod)%,.%/*.go,$(shell go list ./...))

check-dependency-licenses: FORCE prepare-static-check
	@printf "\e[1;36m>> go-licence-detector\e[0m\n"
	@go list -m -mod=readonly -json all | go-licence-detector -includeIndirect -rules .license-scan-rules.json -overrides .license-scan-overrides.jsonl

clean: FORCE
	git clean -dxf build

vars: FORCE
	@printf "BININFO_BUILD_DATE=$(BININFO_BUILD_DATE)\n"
	@printf "BININFO_COMMIT_HASH=$(BININFO_COMMIT_HASH)\n"
	@printf "BININFO_VERSION=$(BININFO_VERSION)\n"
	@printf "DESTDIR=$(DESTDIR)\n"
	@printf "GO_BUILDFLAGS=$(GO_BUILDFLAGS)\n"
	@printf "GO_COVERPKGS=$(GO_COVERPKGS)\n"
	@printf "GO_LDFLAGS=$(GO_LDFLAGS)\n"
	@printf "GO_TESTENV=$(GO_TESTENV)\n"
	@printf "GO_TESTPKGS=$(GO_TESTPKGS)\n"
	@printf "PREFIX=$(PREFIX)\n"
help: FORCE
	@printf "\n"
	@printf "\e[1mUsage:\e[0m\n"
	@printf "  make \e[36m<target>\e[0m\n"
	@printf "\n"
	@printf "\e[1mGeneral\e[0m\n"
	@printf "  \e[36mvars\e[0m                       Display values of relevant Makefile variables.\n"
	@printf "  \e[36mhelp\e[0m                       Display this help.\n"
	@printf "\n"
	@printf "\e[1mBuild\e[0m\n"
	@printf "  \e[36mbuild-all\e[0m                  Build all binaries.\n"
	@printf "  \e[36mbuild/backup-server\e[0m        Build backup-server.\n"
	@printf "  \e[36minstall\e[0m                    Install all binaries. This option understands the conventional 'DESTDIR' and 'PREFIX' environment variables for choosing install locations.\n"
	@printf "\n"
	@printf "\e[1mTest\e[0m\n"
	@printf "  \e[36mcheck\e[0m                      Run the test suite (unit tests and golangci-lint).\n"
	@printf "  \e[36mprepare-static-check\e[0m       Install any tools required by static-check. This is used in CI before dropping privileges, you should probably install all the tools using your package manager\n"
	@printf "  \e[36mrun-golangci-lint\e[0m          Install and run golangci-lint. Installing is used in CI, but you should probably install golangci-lint using your package manager.\n"
	@printf "  \e[36mbuild/cover.out\e[0m            Run tests and generate coverage report.\n"
	@printf "  \e[36mbuild/cover.html\e[0m           Generate an HTML file with source code annotations from the coverage report.\n"
	@printf "  \e[36mstatic-check\e[0m               Run static code checks\n"
	@printf "\n"
	@printf "\e[1mDevelopment\e[0m\n"
	@printf "  \e[36mvendor\e[0m                     Run go mod tidy, go mod verify, and go mod vendor.\n"
	@printf "  \e[36mvendor-compat\e[0m              Same as 'make vendor' but go mod tidy will use '-compat' flag with the Go version from go.mod file as value.\n"
	@printf "  \e[36mlicense-headers\e[0m            Add license headers to all non-vendored .go files.\n"
	@printf "  \e[36mcheck-license-headers\e[0m      Check license headers in all non-vendored .go files.\n"
	@printf "  \e[36mcheck-dependency-licenses\e[0m  Check all dependency licenses using go-licence-detector.\n"
	@printf "  \e[36mclean\e[0m                      Run git clean.\n"

.PHONY: FORCE
