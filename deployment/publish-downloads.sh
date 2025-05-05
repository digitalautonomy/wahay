#!/usr/bin/env bash

set -e

if [ $# -lt 3 ]; then
    echo "usage: $0 <work dir> <web dir> <binary>"
    exit 1
fi

work_dir=$(realpath $1)         # absolute path
web_dir=$(realpath $2)          # absoute path
binary_path=$(realpath $3)      # relative path to the binary in question, either bin/wahay-TAG or bin/wahay-DATE-COMMIT

binary_name=$(basename $binary_path)

TMP_DIR=$work_dir
latest_name="wahay-latest"

sum_file_path=$binary_path.sha256sum
signature_path=$binary_path.sha256sum.asc

if [ -f $sum_file_path ]; then
    binary_sha256_sum=$(grep --only-matching -E "[[:xdigit:]]{64}" $sum_file_path)
    if grep -q $binary_sha256_sum $web_dir/*.sha256sum > /dev/null 2>&1 ; then 
        echo "Binary already exists - not replacing"
        exit 0
    fi
fi

cp $binary_path $sum_file_path $signature_path $web_dir

cd $web_dir

rm -rf $latest_name
rm -rf $latest_name.sha256sum
rm -rf $latest_name.sha256sum.asc

ln -sf $binary_name               wahay-latest
ln -sf $binary_name.sha256sum     wahay-latest.sha256sum
ln -sf $binary_name.sha256sum.asc wahay-latest.sha256sum.asc

rm -rf $web_dir/bundles/$binary_name
mkdir -p $web_dir/bundles/$binary_name
rm -f $web_dir/bundles/latest
cd $web_dir/bundles
ln -sf $binary_name wahay-latest

link_bundle () {
    local path=$1
    local bundle=${path%.tar.bz2}
    local name=$(basename $bundle)

    cp $bundle.tar.bz2 $web_dir/bundles/$binary_name/
    cp $bundle.tar.bz2.sha256sum $web_dir/bundles/$binary_name/
    cp $bundle.tar.bz2.sha256sum.asc $web_dir/bundles/$binary_name/
}

for bundle in $work_dir/*.tar.bz2; do
    link_bundle $bundle
done
