#!/usr/bin/env bash

set -xe

PACKAGE=$1

dep ensure -update $PACKAGE
git checkout vendor/github.com/coyim/gotk3adapter
git checkout vendor/github.com/mxk/go-sqlite/sqlite3