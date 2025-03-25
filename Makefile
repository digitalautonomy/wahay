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

BASE_TAGS := $(GTK_VERSION_TAG),$(GLIB_VERSION_TAG),$(GDK_VERSION_TAG),$(PANGO_VERSION_TAG)
BINARY_TAGS := -tags $(BASE_TAGS),binary
TEST_TAGS := -tags $(BASE_TAGS),test

GIT_VERSION := $(shell git rev-parse HEAD)
GIT_SHORT_VERSION := $(shell git rev-parse --short HEAD)
TAG_VERSION := $(shell git tag -l --contains $$GIT_VERSION | tail -1)
CURRENT_DATE := $(shell TZ='America/Guayaquil' date "+%Y-%m-%d")
BUILD_TIMESTAMP := $(shell TZ='America/Guayaquil' date '+%Y-%m-%d %H:%M:%S')

GOPATH_SINGLE=$(shell echo $${GOPATH%%:*})

BUILD_DIR := bin

PKGS := $(shell go list ./...)
SRC_DIRS := . $(addprefix .,$(subst github.com/digitalautonomy/wahay,,$(PKGS)))
SRC_TEST := $(foreach sdir,$(SRC_DIRS),$(wildcard $(sdir)/*_test.go))
SRC_ALL := $(foreach sdir,$(SRC_DIRS),$(wildcard $(sdir)/*.go))
SRC := $(filter-out $(SRC_TEST), $(SRC_ALL))

SASS_SRC := sass/light-mode/components/*.scss sass/mixins/*.scss sass/light-mode/ui/*.scss sass/utilities/*.scss sass/light-mode/utilities/*.scss sass/variables/*.scss sass/*.scss sass/dark-mode/ui/*.scss sass/dark-mode/utilities/*.scss sass/dark-mode/components/*.scss
LIGHT_CSS_GEN := gui/styles/light-mode-gui.css
DARK_CSS_GEN := gui/styles/dark-mode-gui.css
AUTOGEN := gui/definitions/* gui/styles/* gui/images/* gui/images/help/* gui/config_files/* tor/files/* client/files/*

GO := go
GOBUILD := $(GO) build
GOTEST := $(GO) test
GOINSTALL := $(GO) install

LDFLAGS_VARS := -X 'main.BuildTimestamp=$(BUILD_TIMESTAMP)' -X 'main.BuildCommit=$(GIT_VERSION)' -X 'main.BuildShortCommit=$(GIT_SHORT_VERSION)' -X 'main.BuildTag=$(TAG_VERSION)'
LDFLAGS_REGULAR = -ldflags "$(LDFLAGS_VARS)"
LDFLAGS_WIN = -ldflags "$(LDFLAGS_VARS) -H windowsgui"

COVERPROFILE := coverprofile

export GO111MODULE=on

.PHONY: default check-deps gen-ui-defs deps optional-deps test test-clean coverage coverage-tails build-ci lint gosec ineffassign vet errcheck golangci-lint quality all clean sass-watch build-gui-win

default: build

gen-ui-locale:
	cd gui && make generate-locale

deps-ci:
	go get github.com/modocache/gover
	go install github.com/modocache/gover
	go get github.com/securego/gosec/cmd/gosec
	go install github.com/securego/gosec/cmd/gosec
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_SINGLE)/bin latest

deps: deps-ci
	go install golang.org/x/text/cmd/gotext

test:
	go test -cover -v $(TEST_TAGS) ./...

test-clean: test
	go clean -testcache

coverage:
	$(GOTEST) $(TEST_TAGS) -cover -coverprofile coverlog ./... || true
	$(GO) tool cover -html coverlog
	$(RM) coverlog

$(COVERPROFILE):
	$(GOTEST) -cover -coverprofile $@ ./...

coverage-tails:
	$(GOTEST) $(TEST_TAGS) -cover -coverprofile coverlog ./... || true
	$(GO) tool cover -html coverlog -o ~/Tor\ Browser/coverage.html
	xdg-open ~/Tor\ Browser/coverage.html
	$(RM) coverlog

coverage-dev:
	$(GOTEST) $(TEST_TAGS) -cover -coverprofile coverlog ./... || true
	$(GO) tool cover -html coverlog
	$(RM) coverlog

gui/styles:
	mkdir -p $@

$(LIGHT_CSS_GEN): gui/styles $(SASS_SRC)
	# this is necessary because we have a directory named sass as well, so Make gets confused
	$(shell which sass) ./sass/light-mode-gui.scss $@

$(DARK_CSS_GEN): gui/styles $(SASS_SRC)
	# this is necessary because we have a directory named sass as well, so Make gets confused
	$(shell which sass) ./sass/dark-mode-gui.scss $@

sass-watch: gui/styles $(SASS_SRC)
	# this is necessary because we have a directory named sass as well, so Make gets confused
	$(shell which sass) --watch ./sass/light-mode-gui.scss $(LIGHT_CSS_GEN)
	$(shell which sass) --watch ./sass/dark-mode-gui.scss $(DARK_CSS_GEN)

$(BUILD_DIR)/wahay: $(AUTOGEN) $(SRC)
	go build $(LDFLAGS_REGULAR) $(BINARY_TAGS) -o $(BUILD_DIR)/wahay

$(BUILD_DIR)/wahay.exe: $(AUTOGEN) $(SRC)
	go build $(LDFLAGS_WIN) $(BINARY_TAGS) -o $(BUILD_DIR)/wahay.exe

build: $(BUILD_DIR)/wahay
build-gui-win: $(BUILD_DIR)/wahay.exe

build-ci: $(BUILD_DIR)/wahay
ifeq ($(TAG_VERSION),)
	cp $(BUILD_DIR)/wahay $(BUILD_DIR)/wahay-$(CURRENT_DATE)-$(GIT_SHORT_VERSION)
else
	cp $(BUILD_DIR)/wahay $(BUILD_DIR)/wahay-$(TAG_VERSION)
endif

clean:
	$(RM) -rf $(BUILD_DIR)/wahay
	$(RM) -rf $(BUILD_DIR)/wahay.exe

# QUALITY TOOLS

lint:
	golangci-lint run --disable-all -E golint ./...

gosec:
	gosec -conf .gosec.config.json ./...

ineffassign:
	golangci-lint run --disable-all -E ineffassign ./...

vet:
	golangci-lint run --disable-all -E govet ./...

errcheck:
	golangci-lint run --disable-all -E errcheck ./...

golangci-lint:
	golangci-lint run ./...

quality: golangci-lint
