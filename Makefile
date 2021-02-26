GTK_VERSION=$(shell pkg-config --modversion gtk+-3.0 | tr . _ | cut -d '_' -f 1-2)
GTK_BUILD_TAG="gtk_$(GTK_VERSION)"

GIT_VERSION := $(shell git rev-parse HEAD)
GIT_SHORT_VERSION := $(shell git rev-parse --short HEAD)
TAG_VERSION := $(shell git tag -l --contains $$GIT_VERSION | tail -1)
CURRENT_DATE := $(shell date "+%Y-%m-%d")
BUILD_TIMESTAMP := $(shell TZ='America/Guayaquil' date '+%Y-%m-%d %H:%M:%S')

GOPATH_SINGLE=$(shell echo $${GOPATH%%:*})

BUILD_DIR := bin
BUILD_TOOLS_DIR := .build-tools

PKGS := $(shell go list ./... | grep -v /vendor)
SRC_DIRS := . $(addprefix .,$(subst github.com/digitalautonomy/wahay,,$(PKGS)))
SRC_TEST := $(foreach sdir,$(SRC_DIRS),$(wildcard $(sdir)/*_test.go))
SRC_ALL := $(foreach sdir,$(SRC_DIRS),$(wildcard $(sdir)/*.go))
SRC := $(filter-out $(SRC_TEST), $(SRC_ALL))

GO_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f 1-2)
ifneq ($(GO_VERSION), 1.11)
	SUPPORT_GOSEC = 1
else
	SUPPORT_GOSEC = 0
endif

.PHONY: default check-deps gen-ui-defs deps optional-deps test test-clean run-coverage clean-cover cover cover-ci build-ci lint gosec ineffassign vet errcheck golangci-lint quality all clean

default: build

gen-ui-locale:
	cd gui && make generate-locale

deps-ci:
	go get -u github.com/modocache/gover
	go get -u github.com/rosatolen/esc
ifeq ($(SUPPORT_GOSEC), 1)
	go get -u github.com/securego/gosec/cmd/gosec
endif
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_SINGLE)/bin latest

deps: deps-ci
	go get -u golang.org/x/text/cmd/gotext

optional-deps:
	go get -u github.com/rogpeppe/godef

test:
	go test -cover -v ./client ./config ./gui ./hosting	 ./tor

test-clean: test
	go clean -testcache

run-coverage: clean-cover
	mkdir -p .coverprofiles
	go test -coverprofile=.coverprofiles/client.coverprofile ./client
	go test -coverprofile=.coverprofiles/config.coverprofile ./config
	go test -coverprofile=.coverprofiles/gui.coverprofile ./gui
	go test -coverprofile=.coverprofiles/hosting.coverprofile ./hosting
	go test -coverprofile=.coverprofiles/tor.coverprofile ./tor
	gover .coverprofiles .coverprofiles/gover.coverprofile

clean-cover:
	$(RM) -rf .coverprofiles
	$(RM) -rf coverage.html

cover: run-coverage
	go tool cover -html=.coverprofiles/gover.coverprofile

cover-ci: run-coverage
	go tool cover -html=.coverprofiles/gover.coverprofile -o coverage.html
	go tool cover -func=.coverprofiles/gover.coverprofile

$(BUILD_DIR)/wahay: gui/definitions.go client/gen_client_files.go $(SRC)
	go build -ldflags "-X 'main.BuildTimestamp=$(BUILD_TIMESTAMP)' -X 'main.BuildCommit=$(GIT_VERSION)' -X 'main.BuildShortCommit=$(GIT_SHORT_VERSION)' -X 'main.Build=$(TAG_VERSION)'" -i -tags $(GTK_BUILD_TAG) -o $(BUILD_DIR)/wahay

build: $(BUILD_DIR)/wahay

build-ci: $(BUILD_DIR)/wahay
ifeq ($(TAG_VERSION),)
	cp $(BUILD_DIR)/wahay $(BUILD_DIR)/wahay-$(CURRENT_DATE)-$(GIT_SHORT_VERSION)
else
	cp $(BUILD_DIR)/wahay $(BUILD_DIR)/wahay-$(TAG_VERSION)-$(GIT_SHORT_VERSION)
endif

clean:
	$(RM) -rf $(BUILD_DIR)/wahay
	$(RM) -rf $(BUILD_TOOLS_DIR)

$(BUILD_TOOLS_DIR):
	mkdir -p $@

$(BUILD_TOOLS_DIR)/esc: $(BUILD_TOOLS_DIR)
	@type esc >/dev/null 2>&1 || (echo "The program 'esc' is required but not available. Please install it by running 'make deps'." && exit 1)
	@cp `which esc` $(BUILD_TOOLS_DIR)/esc

client/gen_client_files.go: $(BUILD_TOOLS_DIR)/esc client/files/* client/files/.*
	(cd client; go generate -x client.go)

gui/definitions.go: $(BUILD_TOOLS_DIR)/esc gui/definitions/* gui/styles/* gui/images/* gui/images/help/* gui/config_files/*
	(cd gui; go generate -x ui_reader.go)

# QUALITY TOOLS

lint:
	golangci-lint run --disable-all -E golint ./...

gosec:
ifeq ($(SUPPORT_GOSEC), 1)
	gosec -conf .gosec.config.json ./...
else
	echo '`gosec` is not supported for the current version ($(GO_VERSION)) of `go`';
endif

ineffassign:
	golangci-lint run --disable-all -E ineffassign ./...

vet:
	golangci-lint run --disable-all -E govet ./...

errcheck:
	golangci-lint run --disable-all -E errcheck ./...

golangci-lint:
	golangci-lint run ./...

quality: golangci-lint
