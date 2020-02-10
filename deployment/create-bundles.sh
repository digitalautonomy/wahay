#!/usr/bin/env bash

set -x


# Create working directories
mkdir -p generate-bundles/wahay
mkdir publish-bundles

# Download Binary  Bundles from Nextcloud
cd generate-bundles

echo "$RCLONE_CONFIG" > rclone.conf
rclone copy --config rclone.conf  wahay: .

#extract tor
tar xvf tor-0.4.2.5.tar.bz2 --directory wahay

# Create Ubuntu 18.04 Bundle
tar xvf mumble-ubuntu-18-04.tar.bz2 --directory wahay
cp ../$BINARY_NAME wahay/wahay
tar cjf wahay-bundle-ubuntu-18-04.tar.bz2 wahay
rm -rf wahay/mumble
mv wahay-bundle-ubuntu-18-04.tar.bz2 ../publish-bundles
