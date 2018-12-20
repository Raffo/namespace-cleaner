GOCMD:=go
DEP:=dep
BUILD_DIR:=build
GOBUILD:=$(GOCMD) build
GOCLEAN:=$(GOCMD) clean
GOTEST:=$(GOCMD) test
GOARCH:=amd64
PLATFORMS:=linux
GOOS=$(word 1, $@)
BINARY_NAME=namespace-cleaner
LDFLAGS=-ldflags "-X main.version=$(VERSION)"
VERSION?=$(shell git describe --tags --always --dirty)
KUBERNETES_DEPLOYMENT_TAG?=v4

all: actions deps test build

$(PLATFORMS):
	mkdir -p $(BUILD_DIR)
	GOOS=$(GOOS) GOARCH=$(GOARCH) $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$(GOOS)-$(GOARCH) -v .

build: linux

actions:
	APP_DIR="/go/src/github.com/${GITHUB_REPOSITORY}/" mkdir -p ${APP_DIR} && cp -r ./ ${APP_DIR} && cd ${APP_DIR}

test:
	$(GOTEST) -cover -v ./...

clean:
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)

tag:
	if test "$(TAG)" = "" ; then \
		echo "usage: make tag TAG=1.2.3"; \
		exit 1; \
	fi
	git tag -a $(TAG) -m "$(TAG)"
	git push origin $(TAG)

deps:
	$(GOCMD) get -u github.com/golang/dep/cmd/dep
	$(DEP) ensure

install:
	$(GOCMD) install github.com/Raffo/namespace-cleaner/cmd/...

.PHONY: all build test clean deps tag $(PLATFORMS)
