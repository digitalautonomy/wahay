GTK_VERSION=$(shell pkg-config --modversion gtk+-3.0 | tr . _ | cut -d '_' -f 1-2)
GTK_BUILD_TAG="gtk_$(GTK_VERSION)"

GIT_VERSION=$(shell git rev-parse HEAD)
GIT_SHORT_VERSION=$(shell git rev-parse --short HEAD)
TAG_VERSION=$(shell git tag -l --contains $$GIT_VERSION | tail -1)
CURRENT_DATE=$(shell date "+%Y-%m-%d")

GOPATH_SINGLE=$(shell echo $${GOPATH%%:*})

BUILD_DIR=bin

.PHONY: default check-deps gen-ui-defs deps optional-deps test test-clean run-coverage clean-cover cover cover-ci build build-ci lint gosec ineffassign vet errcheck golangci-lint quality all clean

default: gen-ui-defs build

check-deps:
	@type esc >/dev/null 2>&1 || (echo "The program 'esc' is required but not available. Please install it by running 'make deps'." && exit 1)

gen-ui-defs: check-deps
	cd gui && make generate-ui

gen-ui-locale: check-deps
	cd gui && make generate-locale

gen-client-files: check-deps
	cd client && make

deps:
	go get -u github.com/modocache/gover
	go get -u github.com/rosatolen/esc
	go get -u golang.org/x/text/cmd/gotext
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_SINGLE)/bin v1.21.0

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

build:
	go build -i -tags $(GTK_BUILD_TAG) -o $(BUILD_DIR)/wahay

build-ci:
ifeq ($(TAG_VERSION),)
	go build -i -tags $(GTK_BUILD_TAG) -o $(BUILD_DIR)/wahay-$(CURRENT_DATE)-$(GIT_SHORT_VERSION)
else
	go build -i -tags $(GTK_BUILD_TAG) -o $(BUILD_DIR)/wahay-$(TAG_VERSION)-$(GIT_SHORT_VERSION)
endif

# QUALITY TOOLS

lint:
	golangci-lint run --disable-all -E golint ./...

gosec:
	golangci-lint run --disable-all -E gosec ./...

ineffassign:
	golangci-lint run --disable-all -E ineffassign ./...

vet:
	golangci-lint run --disable-all -E govet ./...

errcheck:
	golangci-lint run --disable-all -E errcheck ./...

golangci-lint:
	golangci-lint run ./...

quality: golangci-lint
