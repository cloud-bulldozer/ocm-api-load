VERSION := v0.4.0
NAME := ocm-load-test
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
	go build -o build/$(NAME) -ldflags "$(LD_FLAGS)" cmd/ocm-load-test.go

build-image: build
	podman build --tag quay.io/cloud-bulldozer/ocm-api-load:$(VERSION) -f Dockerfile

dist: export COPYFILE_DISABLE=1 #teach OSX tar to not put ._* files in tar archive
dist:
	rm -rf build/* release/*
	mkdir -p build release/
	cp automation.py requirements.txt README.md LICENSE build/
	GOOS=linux GOARCH=amd64 go build -o build/$(NAME) -ldflags="$(LD_FLAGS)" cmd/ocm-load-test.go
	tar -C build/ -zcvf $(CURDIR)/release/$(NAME)-linux.tgz automation.py requirements.txt LICENSE README.md $(NAME)
	GOOS=darwin GOARCH=amd64 go build -o build/$(NAME) -ldflags="$(LD_FLAGS)" cmd/ocm-load-test.go
	tar -C build/ -zcvf $(CURDIR)/release/$(NAME)-macos.tgz automation.py requirements.txt LICENSE README.md $(NAME)
	rm build/$(NAME)
	GOOS=windows GOARCH=amd64 go build -o build/$(NAME).exe -ldflags="$(LD_FLAGS)" cmd/ocm-load-test.go
	tar -C build/ -zcvf $(CURDIR)/release/$(NAME)-windows.tgz automation.py requirements.txt LICENSE README.md $(NAME).exe

release: dist
ifndef GITHUB_TOKEN
	$(error GITHUB_TOKEN is undefined)
endif
ifndef GITHUB_USER
	$(error GITHUB_USER is undefined)
endif
	git tag $(VERSION)
	git push origin --tags
	github-release release -u cloud-bulldozer -r ocm-api-load -t $(VERSION) --name $(VERSION)
	github-release upload -u cloud-bulldozer -r ocm-api-load -t $(VERSION) --name $(NAME)-linux.tgz --file release/$(NAME)-linux.tgz
	github-release upload -u cloud-bulldozer -r ocm-api-load -t $(VERSION) --name $(NAME)-macos.tgz --file release/$(NAME)-macos.tgz
	github-release upload -u cloud-bulldozer -r ocm-api-load -t $(VERSION) --name $(NAME)-windows.tgz --file release/$(NAME)-windows.tgz

.PHONY: all build dist release
