#!/usr/bin/env bash

set -x

#Expect to have the new binary files, bundles per distributions and it's sha256sum and signature     files

APP_NAME=wahay
TMP_DIR=~/tmp/deploy_binaries
SUM_FILE_FULL=$(find $TMP_DIR -name "*.sha256sum" | grep -v ".bz2\|.deb\|.rpm" | head -1)
SHA256_SUM_FILE=$(basename $SUM_FILE_FULL)
BINARY_SHA256_SUM=$(grep --only-matching -E "[[:xdigit:]]{64}" $SUM_FILE_FULL)
BINARY_NAME=${SHA256_SUM_FILE%.sha256sum}
BINARY_VERSION=${BINARY_NAME#$APP_NAME-}
SIGNATURE_FILE=$SHA256_SUM_FILE.asc
ALL_BUNDLES=$(find $TMP_DIR -name  "*.bz2" -exec basename {} \;)
ALL_BUNDLES_SHA256_SUM=$(echo "$ALL_BUNDLES"  | sed 's/$/.sha256sum/')
ALL_BUNDLES_SIGNATURES=$(echo "$ALL_BUNDLES"  | sed 's/$/.sha256sum.asc/')
DOWNLOADS_DIR=/usr/local/www/${APP_NAME}/downloads

#Compare NEW_WAHAY_BINARY sha256sum with previous
#hashes to avoid duplicated binaries if a binary 
#is duplicated  clean the $TMP_DIR
grep $BINARY_SHA256_SUM $DOWNLOADS_DIR/*.sha256sum 
if [ $? -eq 0 ] 
then
        echo "Binary already exists"
        rm -f $TMP_DIR/*
        exit 0
fi

#Move binaries to the download page
cd $TMP_DIR
mv $BINARY_NAME $SHA256_SUM_FILE $SIGNATURE_FILE $DOWNLOADS_DIR

#Identified if the file has a date patern, that way we can
#now that is not a tagged version
DATE_FORMAT='20[0-9][0-9]-[0-1][0-9]-[0-3][0-9]'
echo $BINARY_NAME | grep "$DATE_FORMAT"
HAS_DATE=$?

if [ $HAS_DATE -eq 0  ] 
then
        cd $DOWNLOADS_DIR
        
        if [ -L ${APP_NAME}-latest ] ; then
                rm ${APP_NAME}-latest*
        fi

        ln -s $BINARY_NAME ${APP_NAME}-latest
        ln -s $SHA256_SUM_FILE ${APP_NAME}-latest.sha256sum
        ln -s $SIGNATURE_FILE ${APP_NAME}-latest.sha256sum.asc
else
 
        #Retrieve WAHAY_TAG name        
        WAHAY_TAG_NAME=$(echo ${BINARY_NAME%-*******})

        cd $DOWNLOADS_DIR
  
        if [ -L $WAHAY_TAG_NAME ] ; then
                rm $WAHAY_TAG_NAME*
        fi
    
        ln -s $BINARY_NAME $WAHAY_TAG_NAME
        ln -s $SHA256_SUM_FILE $WAHAY_TAG_NAME.sha256sum
        ln -s $SIGNATURE_FILE $WAHAY_TAG_NAME.sha256sum.asc

fi

#Move bundles to the download directory
cd $TMP_DIR
mkdir -p $DOWNLOADS_DIR/bundles/$BINARY_NAME
echo "$ALL_BUNDLES" | xargs -I file mv file $DOWNLOADS_DIR/bundles/$BINARY_NAME
echo "$ALL_BUNDLES_SHA256_SUM" | xargs -I file mv file $DOWNLOADS_DIR/bundles/$BINARY_NAME
echo "$ALL_BUNDLES_SIGNATURES" | xargs -I file mv file $DOWNLOADS_DIR/bundles/$BINARY_NAME

# Create latest and tag name symlinks for linux bundles
if [ $HAS_DATE -eq 0  ]
then
        cd $DOWNLOADS_DIR
        rm -f *latest.tar.bz2*
	
	echo "$ALL_BUNDLES" |  cut -d "-" -f 1,2,3 | xargs -I file find bundles/$BINARY_NAME -name "file*.bz2" -exec ln -s {} file-latest.tar.bz2 \;
	echo "$ALL_BUNDLES" |  cut -d "-" -f 1,2,3 | xargs -I file find bundles/$BINARY_NAME -name "file*.sha256sum" -exec ln -s {} file-latest.tar.bz2.sha256sum \;
	echo "$ALL_BUNDLES" |  cut -d "-" -f 1,2,3 | xargs -I file find bundles/$BINARY_NAME -name "file*.asc" -exec ln -s {} file-latest.tar.bz2.sha256sum.asc \;
else

        #Retrieve WAHAY_TAG name
        WAHAY_TAG_NAME=${BINARY_NAME%-*******}

        cd $DOWNLOADS_DIR
        rm  -rf $WAHAY_TAG_NAME*bz2*

	echo "$ALL_BUNDLES" |  cut -d "-" -f 2,3 | xargs -I file find bundles/$BINARY_NAME -name "*file*.bz2" -exec ln -s {} $WAHAY_TAG_NAME-file.bz2 \;
	echo "$ALL_BUNDLES" |  cut -d "-" -f 2,3 | xargs -I file find bundles/$BINARY_NAME -name "*file*.bz2.sha256sum" -exec ln -s {} $WAHAY_TAG_NAME-file.bz2.sha256sum \;
	echo "$ALL_BUNDLES" |  cut -d "-" -f 2,3 | xargs -I file find bundles/$BINARY_NAME -name "*file*.bz2.sha256sum.asc" -exec ln -s {} $WAHAY_TAG_NAME-file.bz2.sha256sum.asc \;
fi

#Move Linux Packages to the Download directory
cd $TMP_DIR
mkdir -p $DOWNLOADS_DIR/linux-packages/$BINARY_NAME
mv wahay-ubuntu-${BINARY_VERSION}-amd64.deb* $DOWNLOADS_DIR/linux-packages/$BINARY_NAME/

#Create symlinks for Linux Packagres (just Ubuntu for now)
if [ $HAS_DATE -eq 0  ]
then
        cd $DOWNLOADS_DIR
        rm -f *latest.deb*

	ln -s linux-packages/$BINARY_NAME/${APP_NAME}-ubuntu-$BINARY_VERSION-amd64.deb ${APP_NAME}-ubuntu-latest.deb
	ln -s linux-packages/$BINARY_NAME/${APP_NAME}-ubuntu-$BINARY_VERSION-amd64.deb.sha256sum ${APP_NAME}-ubuntu-latest.deb.sha256sum
	ln -s linux-packages/$BINARY_NAME/${APP_NAME}-ubuntu-$BINARY_VERSION-amd64.deb.sha256sum.asc ${APP_NAME}-ubuntu-latest.deb.sha256sum.asc
else

        #Retrieve WAHAY_TAG name
        WAHAY_TAG_NAME=${BINARY_NAME%-*******}

        cd $DOWNLOADS_DIR
        rm  -rf $WAHAY_TAG_NAME*deb*

	ln -s linux-packages/$BINARY_NAME/${APP_NAME}-ubuntu-$BINARY_VERSION-amd64.deb ${APP_NAME}-ubuntu-${WAHAY_TAG_NAME}.deb
        ln -s linux-packages/$BINARY_NAME/${APP_NAME}-ubuntu-$BINARY_VERSION-amd64.deb.sha256sum ${APP_NAME}-ubuntu-${WAHAY_TAG_NAME}.deb.sha256sum
        ln -s linux-packages/$BINARY_NAME/${APP_NAME}-ubuntu-$BINARY_VERSION-amd64.deb.sha256sum.asc ${APP_NAME}-ubuntu-${WAHAY_TAG_NAME}.deb.sha256sum.asc
fi
