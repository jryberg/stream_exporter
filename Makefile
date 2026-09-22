DOCKER ?= docker

CROSSBUILD_PLATFORMS = linux/amd64 linux/arm64 windows/amd64 darwin/amd64 darwin/arm64

VERSION   := $(shell git describe --tags --dirty=-dirty)
REVISION  := $(shell git describe --abbrev=0 --always --match=always-commit-hash --dirty=-dirty)
BRANCH    := $(or $(shell git symbolic-ref --short HEAD 2>/dev/null), "from-release-tag")
BUILDDATE := $(shell date --iso-8601=seconds)
BUILDUSER ?= $(USER)
BUILDHOST ?= $(HOSTNAME)
LDFLAGS    = -X github.com/prometheus/common/version.Version=$(VERSION) \
             -X github.com/prometheus/common/version.Revision=$(REVISION) \
             -X github.com/prometheus/common/version.Branch=$(BRANCH) \
             -X github.com/prometheus/common/version.BuildUser=$(BUILDUSER)@$(BUILDHOST) \
             -X github.com/prometheus/common/version.BuildDate=$(BUILDDATE)

all: build test

test:
	@echo ">> testing code"
	@go test -v -cover ./...

build:
	@echo ">> building binaries"
	@go build -ldflags="$(LDFLAGS)"

crossbuild:
	@echo ">> cross-building"
	@for platform in $(CROSSBUILD_PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; ext=; \
		if [ "$${os}" = "windows" ]; then ext=.exe; fi; \
		CGO_ENABLED=0 GOOS=$${os} GOARCH=$${arch} go build -trimpath -ldflags="-s $(LDFLAGS)" \
			-o "binaries/stream_exporter_$${os}_$${arch}$${ext}" . || exit 1; \
	done

release:
	@echo ">> uploading release ${VERSION}"
	@gh release upload ${VERSION} binaries/*

docker:
	@echo ">> building docker image"
	@$(DOCKER) build -t jryberg/stream_exporter .

.PHONY: all build crossbuild test release docker
