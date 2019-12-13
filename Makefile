GTK_VERSION=$(shell pkg-config --modversion gtk+-3.0 | tr . _ | cut -d '_' -f 1-2)
GTK_BUILD_TAG="gtk_$(GTK_VERSION)"

GIT_VERSION=$(shell git rev-parse HEAD)
TAG_VERSION=$(shell git tag -l --contains $$GIT_VERSION | tail -1)

GOPATH_SINGLE=$(shell echo $${GOPATH%%:*})

BUILD_DIR=bin

default: gen-ui-defs build

check-deps:
	@type esc >/dev/null 2>&1 || (echo "The program 'esc' is required but not available. Please install it by running 'make deps'." && exit 1)

gen-ui-defs: check-deps
	make -C gui

deps:
	go get -u github.com/modocache/gover
	go get -u github.com/rosatolen/esc
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOPATH_SINGLE)/bin v1.21.0

optional-deps:
	go get -u github.com/rogpeppe/godef

test:
	go test -cover -v ./config ./gui

run-coverage: clean-cover
	mkdir -p .coverprofiles
	go test -coverprofile=.coverprofiles/config.coverprofile ./config
	go test -coverprofile=.coverprofiles/gui.coverprofile ./gui
	gover .coverprofiles .coverprofiles/gover.coverprofile

clean-cover:
	$(RM) -rf .coverprofiles
	$(RM) -rf coverage.html

cover: run-coverage
	go tool cover -html=.coverprofiles/gover.coverprofile

cover-ci: run-coverage
	go tool cover -html=.coverprofiles/gover.coverprofile -o coverage.html

build:
	go build -i -tags $(GTK_BUILD_TAG) -o $(BUILD_DIR)/tonio


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
