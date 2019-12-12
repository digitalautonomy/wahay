GIT_VERSION=$(shell git rev-parse HEAD)
TAG_VERSION=$(shell git tag -l --contains $$GIT_VERSION | tail -1)

BUILD_DIR=bin

deps:
	go get -u github.com/modocache/gover

test:
	go test -cover -v ./...

run-coverage: clean-cover
	mkdir -p .coverprofiles
	go test -coverprofile=.coverprofiles/main.coverprofile
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
	go build -i -o $(BUILD_DIR)/tonio
