export PROJECT_NAME ?= qp
SHELL := /bin/bash -o pipefail
VERSION ?=`git describe --tags`
DATE=`date -u +"%Y-%m-%dT%H:%M:%SZ"`
VERSION_PACKAGE = github.com/queryplan-ai/qp/pkg/version

BIN_DIR := $(shell pwd)/bin
export PATH := $(BIN_DIR):$(PATH)

export CGO_ENABLED=0

GIT_TREE = $(shell git rev-parse --is-inside-work-tree 2>/dev/null)
ifneq "$(GIT_TREE)" ""
define GIT_UPDATE_INDEX_CMD
git update-index --assume-unchanged
endef
define GIT_SHA
`git rev-parse HEAD`
endef
else
define GIT_UPDATE_INDEX_CMD
echo "Not a git repo, skipping git update-index"
endef
define GIT_SHA
""
endef
endif

BUILD_GO_FLAGS :=

define LDFLAGS
-ldflags "\
	-extldflags \"-static\" \
	-X ${VERSION_PACKAGE}.version=${VERSION} \
	-X ${VERSION_PACKAGE}.gitSHA=${GIT_SHA} \
	-X ${VERSION_PACKAGE}.buildTime=${DATE} \
"
endef

define LDFLAGS_RELEASE
-ldflags "\
	-s -w \
	-extldflags \"-static\" \
	-X ${VERSION_PACKAGE}.version=${VERSION} \
	-X ${VERSION_PACKAGE}.gitSHA=${GIT_SHA} \
	-X ${VERSION_PACKAGE}.buildTime=${DATE} \
"
endef

CURRENT_USER := $(shell id -u -n)
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org

.PHONE: all
all: build

deps:
	go mod download

.PHONY: build
build:
	go build -a -tags netgo $(BUILD_GO_FLAGS) $(LDFLAGS) -o bin/qp ./cmd/qp

.PHONY: release
release:
	go build -a -tags netgo $(BUILD_GO_FLAGS) $(LDFLAGS_RELEASE) -o bin/qp ./cmd/qp


.PHONY: fmt
fmt:
	go fmt ./pkg/... ./cmd/...

.PHONY: vet
vet:
	go vet ./pkg/... ./cmd/...

.PHONY: test
test: vet
	go test ./pkg/... ./cmd/...

