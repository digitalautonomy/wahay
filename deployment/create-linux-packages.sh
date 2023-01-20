#!/usr/bin/env bash

APP_NAME=wahay
BINARY_BASE_NAME=$(basename $BINARY_NAME)
BINARY_VERSION=${BINARY_BASE_NAME#$APP_NAME-}

mkdir -p publish-linux-packages

cd packaging
make ubuntu-package-ci

cd ../publish-linux-packages
sha256sum ${APP_NAME}-ubuntu-${BINARY_VERSION}-amd64.deb > ${APP_NAME}-ubuntu-${BINARY_VERSION}-amd64.deb.sha256sum
gpg --detach-sign --armor -u 3EE89711B35F8B3089646FCBF3B1159FC97D5490 ${APP_NAME}-ubuntu-${BINARY_VERSION}-amd64.deb.sha256sum
