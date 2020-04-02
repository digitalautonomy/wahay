#!/usr/bin/env bash

set -x

DISTRO_FILE="../deployment/supported-bundle-distros.txt"
APP_NAME=wahay
BINARY_BASE_NAME=$(basename $BINARY_NAME)
BINARY_VERSION=${BINARY_BASE_NAME#$APP_NAME}

# Create working directories
mkdir generate-bundles
mkdir publish-bundles

# Download Binary  Bundles from Nextcloud
cd generate-bundles

echo "$RCLONE_CONFIG" > rclone.conf
rclone copy --config rclone.conf  $APP_NAME:$APP_NAME-bundles .

#extract tor
tar xf tor-0.4.2.5.tar.bz2

pwd

ls ..

# Create distro bundles
while IFS= read -r DISTRO_NAME
do
 DISTRO_DIR=${APP_NAME}-${DISTRO_NAME}${BINARY_VERSION}
 mkdir $DISTRO_DIR
 tar xf mumble-${DISTRO_NAME}.tar.bz2 --directory $DISTRO_DIR
 cp -r tor $DISTRO_DIR
 cp -r ../packaging/bundles/* $DISTRO_DIR
 cp ../$BINARY_NAME $DISTRO_DIR
 cd $DISTRO_DIR
 ln -s $BINARY_BASE_NAME $APP_NAME
 cd ..
 tar cjf wahay-${DISTRO_NAME}$BINARY_VERSION.tar.bz2 $DISTRO_DIR
 sha256sum wahay-${DISTRO_NAME}$BINARY_VERSION.tar.bz2 > wahay-${DISTRO_NAME}$BINARY_VERSION.tar.bz2.sha256sum
 gpg --detach-sign --armor -u 01242FFAB8CE1EC0C8F54456A8854162D28F171E wahay-${DISTRO_NAME}$BINARY_VERSION.tar.bz2.sha256sum
 mv wahay-${DISTRO_NAME}$BINARY_VERSION.tar.bz2* ../publish-bundles
done < "$DISTRO_FILE"

