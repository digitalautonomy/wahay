GLIB_VERSION := $(shell pkg-config --modversion glib-2.0 | tr . _ | cut -d '_' -f 1-2)
GLIB_VERSION_TAG := "glib_$(GLIB_VERSION)"

GTK_VERSION_FULL=$(shell pkg-config --modversion gtk+-3.0)
GTK_VERSION_PATCH=$(shell echo $(GTK_VERSION_FULL) | cut -f3 -d.)
GTK_VERSION=$(shell echo $(GTK_VERSION_FULL) | tr . _ | cut -d '_' -f 1-2)
GTK_VERSION_TAG="gtk_$(GTK_VERSION)"

# All this is necessary to downgrade the gtk version used to 3.22 if the
# 3.24 patch level is lower than 14. The reason for that is that
# a new variable was introduced at 3.24.14, and older patch levels
# won't compile with gotk3

GTK_VERSION_PATCH_LESS14=$(shell expr $(GTK_VERSION_PATCH) \< 14)
ifeq ($(GTK_VERSION_TAG),"gtk_3_24")
ifeq ($(GTK_VERSION_PATCH_LESS14),1)
GTK_VERSION_TAG="gtk_3_22"
endif
endif

GDK_VERSION := $(shell pkg-config --modversion gdk-3.0 | tr . _ | cut -d '_' -f 1-2)
GDK_VERSION_TAG := "gdk_$(GDK_VERSION)"

PANGO_VERSION := $(shell pkg-config --modversion pango | tr . _ | cut -d '_' -f 1-2)
PANGO_VERSION_TAG := "pango_$(PANGO_VERSION)"

BINARY_TAGS := -tags $(GTK_VERSION_TAG),$(GLIB_VERSION_TAG),$(GDK_VERSION_TAG),$(PANGO_VERSION_TAG),binary

GIT_VERSION := $(shell git rev-parse HEAD)
GIT_SHORT_VERSION := $(shell git rev-parse --short HEAD)
TAG_VERSION := $(shell git tag -l --contains $$GIT_VERSION | tail -1)
CURRENT_DATE := $(shell TZ='America/Guayaquil' date "+%Y-%m-%d")
BUILD_TIMESTAMP := $(shell TZ='America/Guayaquil' date '+%Y-%m-%d %H:%M:%S')

GOPATH_SINGLE=$(shell echo $${GOPATH%%:*})

BUILD_DIR := bin
BUILD_TOOLS_DIR := .build-tools

PKGS := $(shell go list ./...)
SRC_DIRS := . $(addprefix .,$(subst github.com/digitalautonomy/wahay,,$(PKGS)))
SRC_TEST := $(foreach sdir,$(SRC_DIRS),$(wildcard $(sdir)/*_test.go))
SRC_ALL := $(foreach sdir,$(SRC_DIRS),$(wildcard $(sdir)/*.go))
SRC := $(filter-out $(SRC_TEST), $(SRC_ALL))

OLDEST_COMPATIBLE_GOLANG_VERSION := 1.15
OLDEST_COMPATIBLE_GOLANG_VERSION_MAJOR := $(shell echo $(OLDEST_COMPATIBLE_GOLANG_VERSION) | cut -f1 -d.)
OLDEST_COMPATIBLE_GOLANG_VERSION_MINOR := $(shell echo $(OLDEST_COMPATIBLE_GOLANG_VERSION) | cut -f2 -d.)

NEWEST_COMPATIBLE_GOLANG_VERSION := 1.19
NEWEST_COMPATIBLE_GOLANG_VERSION_MAJOR := $(shell echo $(NEWEST_COMPATIBLE_GOLANG_VERSION) | cut -f1 -d.)
NEWEST_COMPATIBLE_GOLANG_VERSION_MINOR := $(shell echo $(NEWEST_COMPATIBLE_GOLANG_VERSION) | cut -f2 -d.)

GO_VERSION = $(shell go version | cut -c 14- | cut -d' ' -f1 | cut -d'.' -f 1-2)
GO_VERSION_MAJOR := $(shell echo $(GO_VERSION) | cut -f1 -d.)
GO_VERSION_MINOR := $(shell echo $(GO_VERSION) | cut -f2 -d.)

ifeq ($(GO_VERSION_MAJOR), 1)
	GO_OLDER_THAN_COMPATIBLE=$(shell echo $(GO_VERSION_MINOR)\<$(OLDEST_COMPATIBLE_GOLANG_VERSION_MINOR) | bc )
	GO_NEWER_THAN_COMPATIBLE=$(shell echo $(GO_VERSION_MINOR)\>$(NEWEST_COMPATIBLE_GOLANG_VERSION_MINOR) | bc )
else
	GO_OLDER_THAN_COMPATIBLE=$(shell echo $(GO_VERSION_MAJOR)\<$(OLDEST_COMPATIBLE_GOLANG_VERSION_MAJOR) | bc )
	GO_NEWER_THAN_COMPATIBLE=$(shell echo $(GO_VERSION_MAJOR)\>$(NEWEST_COMPATIBLE_GOLANG_VERSION_MAJOR) | bc )
endif

GOSEC_COMPATIBILITY_VERSION := 1.19
GOLANGCI_COMPATIBILITY_VERSION := 1.19

GOSEC_COMPARE = $(shell ./check_version.rb $(GOSEC_COMPATIBILITY_VERSION) $(GO_VERSION))
GOLANGCI_COMPARE = $(shell ./check_version.rb $(GOLANGCI_COMPATIBILITY_VERSION) $(GO_VERSION))

ifneq ($(GOSEC_COMPARE), gt)
	SUPPORT_GOSEC=1
else
	SUPPORT_GOSEC=0
endif

ifneq ($(GOLANGCI_COMPARE), gt)
	SUPPORT_GOLANGCI=1
else
	SUPPORT_GOLANGCI=0
endif

GO := go
GOBUILD := $(GO) build
GOTEST := $(GO) test
GOINSTALL := $(GO) install

COVERPROFILE := coverprofile

export GO111MODULE=on

.PHONY: default check-version check-deps gen-ui-defs deps optional-deps test test-clean coverage coverage-tails build-ci lint gosec ineffassign vet errcheck golangci-lint quality all clean

default: build

check-version:
ifeq ($(GO_OLDER_THAN_COMPATIBLE),1)
	@echo "Your version of Golang is too old to be compatible - the oldest version supported is $(OLDEST_COMPATIBLE_GOLANG_VERSION)"
	@exit 1
else
ifeq ($(GO_NEWER_THAN_COMPATIBLE),1)
	@echo "Your version of Golang is too new to be compatible - the newest version supported is $(NEWEST_COMPATIBLE_GOLANG_VERSION)"
	@exit 1
endif
endif

gen-ui-locale:
	cd gui && make generate-locale

deps-ci: check-version
ifeq ($(SUPPORT_GOVER), 1)
	go get github.com/modocache/gover
	go install github.com/modocache/gover
endif
	go get github.com/rosatolen/esc
	go install github.com/rosatolen/esc
ifeq ($(SUPPORT_GOSEC), 1)
	go get github.com/securego/gosec/cmd/gosec
	go install github.com/securego/gosec/cmd/gosec
endif
ifeq ($(SUPPORT_GOLANGCI), 1)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_SINGLE)/bin latest
endif

deps: deps-ci
	go install golang.org/x/text/cmd/gotext

test: check-version
	go test -cover -v ./client ./config ./gui ./hosting	 ./tor

test-clean: test
	go clean -testcache

coverage:
	$(GOTEST) $(BINARY_TAGS) -cover -coverprofile coverlog ./... || true
	$(GO) tool cover -html coverlog
	$(RM) coverlog

$(COVERPROFILE):
	$(GOTEST) -cover -coverprofile $@ ./...

coverage-tails:
	$(GOTEST) $(BINARY_TAGS) -cover -coverprofile coverlog ./... || true
	$(GO) tool cover -html coverlog -o ~/Tor\ Browser/coverage.html
	xdg-open ~/Tor\ Browser/coverage.html
	$(RM) coverlog

$(BUILD_DIR)/wahay: check-version gui/definitions.go client/gen_client_files.go $(SRC)
	go build -ldflags "-X 'main.BuildTimestamp=$(BUILD_TIMESTAMP)' -X 'main.BuildCommit=$(GIT_VERSION)' -X 'main.BuildShortCommit=$(GIT_SHORT_VERSION)' -X 'main.Build=$(TAG_VERSION)'" $(BINARY_TAGS) -o $(BUILD_DIR)/wahay

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
	./build/find_esc.sh $(BUILD_TOOLS_DIR)

client/gen_client_files.go: $(BUILD_TOOLS_DIR)/esc client/files/* client/files/.*
	(cd client; go generate -x client.go)

gui/definitions.go: $(BUILD_TOOLS_DIR)/esc gui/definitions/* gui/styles/* gui/images/* gui/images/help/* gui/config_files/*
	(cd gui; go generate -x ui_reader.go)

# QUALITY TOOLS

lint: check-version
ifeq ($(SUPPORT_GOLANGCI), 1)
	golangci-lint run --disable-all -E golint ./...
else
	echo '`golangci` is not supported for the current version ($(GO_VERSION)) of `go`';
endif

gosec: check-version
ifeq ($(SUPPORT_GOSEC), 1)
	gosec -conf .gosec.config.json ./...
else
	echo '`gosec` is not supported for the current version ($(GO_VERSION)) of `go`';
endif

ineffassign: check-version
ifeq ($(SUPPORT_GOLANGCI), 1)
	golangci-lint run --disable-all -E ineffassign ./...
else
	echo '`golangci` is not supported for the current version ($(GO_VERSION)) of `go`';
endif

vet: check-version
ifeq ($(SUPPORT_GOLANGCI), 1)
	golangci-lint run --disable-all -E govet ./...
else
	echo '`golangci` is not supported for the current version ($(GO_VERSION)) of `go`';
endif

errcheck: check-version
ifeq ($(SUPPORT_GOLANGCI), 1)
	golangci-lint run --disable-all -E errcheck ./...
else
	echo '`golangci` is not supported for the current version ($(GO_VERSION)) of `go`';
endif

golangci-lint: check-version
ifeq ($(SUPPORT_GOLANGCI), 1)
	golangci-lint run ./...
else
	echo '`golangci` is not supported for the current version ($(GO_VERSION)) of `go`';
endif

quality: golangci-lint
