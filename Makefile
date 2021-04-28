VERSION := v0.1.0
# Enable Go modules:
export GO111MODULE=on
export GOPROXY=https://proxy.golang.org
export GOPRIVATE=gitlab.cee.redhat.com/service

# Disable CGO so that we always generate static binaries:
export CGO_ENABLED=0

# LD Flags
DATE := $(shell date -u +%Y%m%d.%H%M%S)
COMMIT_ID := $(shell git rev-parse --short HEAD)
GIT_REPO := $(shell git config --get remote.origin.url)
# Go tools flags
LD_FLAGS := -X github.com/cloud-bulldozer/ocm-api-load/pkg/cmd.BuildVersion=$(VERSION)
LD_FLAGS += -X github.com/cloud-bulldozer/ocm-api-load/pkg/cmd.BuildCommit=$(COMMIT_ID)
LD_FLAGS += -X github.com/cloud-bulldozer/ocm-api-load/pkg/cmd.BuildDate=$(DATE)

all: build

build:
	go build -ldflags "$(LD_FLAGS)" cmd/ocm-load-test.go

.PHONY: all build