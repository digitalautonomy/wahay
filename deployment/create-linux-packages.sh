#!/usr/bin/env bash

mkdir -p publish-linux-packages

cd packaging
make ubuntu-package-ci


