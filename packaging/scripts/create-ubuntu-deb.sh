#!/usr/bin/env bash

CURRENT_VERSION=$(echo 0~$(date "+%y%m%d%H%M%S"))

if [ $1 == "local"  ]
then
	cp ../bin/wahay ubuntu/ubuntu/usr/bin
	sed "s/##VERSION##/$CURRENT_VERSION/g" ubuntu/templates/control > ubuntu/ubuntu/DEBIAN/control
	fakeroot dpkg-deb --build ubuntu/ubuntu ../bin/wahay_${CURRENT_VERSION}_amd64.deb
elif [ $1 == "ci" ]
then
	BINARY_BASE_NAME=$(basename $BINARY_NAME)
	cp ../$BINARY_NAME ubuntu/ubuntu/usr/bin
	sed "s/##VERSION##/$CURRENT_VERSION/g" ubuntu/templates/control > ubuntu/ubuntu/DEBIAN/control
	cd  ubuntu/ubuntu/usr/bin 
	rm -f wahay*
	ln -s $BINARY_BASE_NAME wahay
	cd -
	fakeroot dpkg-deb --build ubuntu/ubuntu ../publish-linux-packages/wahay_${CURRENT_VERSION}_amd64.deb
else
	echo "Unknow argument value"

fi
