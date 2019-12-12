GTK_VERSION=$(shell pkg-config --modversion gtk+-3.0 | tr . _ | cut -d '_' -f 1-2)
GTK_BUILD_TAG="gtk_$(GTK_VERSION)"

GIT_VERSION=$(shell git rev-parse HEAD)
TAG_VERSION=$(shell git tag -l --contains $$GIT_VERSION | tail -1)

BUILD_DIR=bin

deps:
	go get -u github.com/modocache/gover

test:
	go test -cover -tags $(GTK_BUILD_TAG) -v ./...

run-coverage: clean-cover
	mkdir -p .coverprofiles
	go test -tags $(GTK_BUILD_TAG) -coverprofile=.coverprofiles/main.coverprofile
	go test -tags $(GTK_BUILD_TAG) -coverprofile=.coverprofiles/config.coverprofile ./config
	go test -tags $(GTK_BUILD_TAG) -coverprofile=.coverprofiles/gui.coverprofile ./gui
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
