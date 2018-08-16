#!/bin/bash


set -x;

SVN=http://cld2.googlecode.com/svn/trunk/
DIR=cld2
# Checkout a temporary copy of the repository
svn co ${SVN} ${DIR}

cp ${DIR}/internal/*.h .
cp ${DIR}/public/*.h .


# Use the script for compiling the lib to gather the files
# for the Go package.
cd ${DIR}/internal

calls=0

function g++ {
    # First time g++ is called it builds the "non-full" lib
    if [ $calls -lt 1 ]; then
        ((calls+=1))
        return;
    fi

    while (( "$#" )); do
        if [ "$1" = '-o' ]; then
            break;
        fi
        if [[ $1 == -* ]]; then
            shift;
            continue;
        fi
        cp $1 ../../;
        shift;
    done
}

source ./compile_libs.sh

cd -

sed -i 's/..\/public\///' *.h *.cc
sed -i 's/..\/internal\///' *.h

rm -fr ${DIR}
