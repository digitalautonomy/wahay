#!/usr/bin/env bash

set -x

#Expect to have the new binary, sha256sum and signature
TMP_DIR=~/tmp/deploy_binaries

SUM_FILE_FULL=$(find $TMP_DIR -name '*.sha256sum' | head -1)

SHA256_SUM_FILE=$(basename $SUM_FILE_FULL)

BINARY_SHA256_SUM=$(grep --only-matching -E "[[:xdigit:]]{64}" $SUM_FILE_FULL)

BINARY_NAME=${SHA256_SUM_FILE%.sha256sum}

SIGNATURE_FILE=$SHA256_SUM_FILE.asc

DOWNLOADS_DIR=/usr/local/www/tonio/downloads

WEBSITE_DOCUMENT_ROOT=/usr/local/www/tonio/

#Compare NEW_TONIO_BINARY sha256sum with previous
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
mv $TMP_DIR/tonio* $DOWNLOADS_DIR

#Identified if the file has a date patern, that way we can
#now that is not a tagged version
DATE_FORMAT='-20[0-9]{2}-[0-1][0-9]-[0-3][0-9]-'

echo $BINARY_NAME | grep "$DATE_FORMAT"

if [ $? -eq 0  ] 
then
        cd $DOWNLOADS_DIR
        
        if [ -L tonio-latest ] ; then
                rm tonio-latest*
        fi

        ln -s $BINARY_NAME tonio-latest
        ln -s $SHA256_SUM_FILE tonio-latest.sha256sum
        ln -s $SIGNATURE_FILE tonio-latest.sha256sum.asc
else
 
        #Retrieve TONIO_TAG name        
        TONIO_TAG_NAME=$(echo ${BINARY_NAME%-*******})

        cd $DOWNLOADS_DIR
  
        if [ -L $TONIO_TAG_NAME ] ; then
                rm $TONIO_TAG_NAME*
        fi
    
        ln -s $BINARY_NAME $TONIO_TAG_NAME
        ln -s $SHA256_SUM_FILE $TONIO_TAG_NAME.sha256sum
        ln -s $SIGNATURE_FILE $TONIO_TAG_NAME.sha256sum.asc


fi

#Generate website
~/bin/generate-downloads-html.sh $DOWNLOADS_DIR > $WEBSITE_DOCUMENT_ROOT/downloads.html
