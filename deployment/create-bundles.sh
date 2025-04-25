#!/usr/bin/env bash

# The distribution of Tor to download for bundling comes from. It is the "tor expert bundle"
#   - https://www.torproject.org/download/tor/
# The different bundles come as artifacts from the projects/wahay/bundles project in Gitlab
# 

set -xe

export WAHAY_VERSION=0.2
export TOR_VERSION=0.4.7.13

DISTRO_FILE="../deployment/supported-bundle-distros.txt"
BINARY_BASE_NAME=$(basename $BINARY_NAME)

# Create working directories
mkdir generate-bundles
mkdir publish-bundles

cd generate-bundles
tar xf ../tmp_bundle/linux-x86_64-tor-$TOR_VERSION.tar.gz

# Create distro bundles
while IFS= read -r DISTRO_NAME
do
 DISTRO_DIR=wahay-$DISTRO_NAME-$WAHAY_VERSION
 mkdir $DISTRO_DIR
 tar xf ../tmp_bundle/mumble-$DISTRO_NAME-wahay-$WAHAY_VERSION.tar.bz2 --directory $DISTRO_DIR
 cp -r tor $DISTRO_DIR
 cp -r ../packaging/bundles/* $DISTRO_DIR
 cp ../$BINARY_NAME $DISTRO_DIR
 cd $DISTRO_DIR
 ln -sf $BINARY_BASE_NAME wahay
 cd ..
 tar cjf wahay-$DISTRO_NAME-$WAHAY_VERSION.tar.bz2 $DISTRO_DIR
 sha256sum wahay-$DISTRO_NAME-$WAHAY_VERSION.tar.bz2 > wahay-$DISTRO_NAME-$WAHAY_VERSION.tar.bz2.sha256sum
 gpg --batch --detach-sign --armor -u 3EE89711B35F8B3089646FCBF3B1159FC97D5490 wahay-$DISTRO_NAME-$WAHAY_VERSION.tar.bz2.sha256sum
 mv wahay-$DISTRO_NAME-$WAHAY_VERSION.tar.bz2* ../publish-bundles
done < "$DISTRO_FILE"
