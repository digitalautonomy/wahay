#!/usr/bin/env bash

mkdir -p publish-linux-packages

cd packaging
make ubuntu-package-ci

#print Ruby version for debugging
ruby --version
/usr/bin/ruby --version
