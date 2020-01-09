#!/usr/bin/env bash

set -xe

PACKAGE=$1

dep ensure -add $PACKAGE
git checkout vendor/github.com/coyim/gotk3adapter
