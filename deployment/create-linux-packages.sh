#!/usr/bin/env bash

APP_NAME=wahay
BINARY_BASE_NAME=$(basename $BINARY_NAME)
BINARY_VERSION=${BINARY_BASE_NAME#$APP_NAME-}

mkdir -p publish-linux-packages

cd packaging
make ubuntu-package-ci

cd ../publish-linux-packages
sha256sum ${APP_NAME}-ubuntu-${BINARY_VERSION}-amd64.deb > ${APP_NAME}-ubuntu-${BINARY_VERSION}-amd64.deb.sha256sum
gpg --detach-sign --armor -u 01242FFAB8CE1EC0C8F54456A8854162D28F171E ${APP_NAME}-ubuntu-${BINARY_VERSION}-amd64.deb.sha256sum
