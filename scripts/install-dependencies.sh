#!/usr/bin/env sh

GLIDEBIN=$GOPATH/bin/glide
GLIDE_CACHE=~/glide
GLIDE_BIN_CACHED=$GLIDE_CACHE/glide
GETH_BIN_CACHED=/tmp/geth_bin/geth

if [-e "${GLIDE_BIN_CACHED}"]
then
    echo "Found existing glide binary in ${GLIDE_BIN_CACHED}. Not downloading again - just linking."
    ln -s $GLIDE_BIN_CACHED $GLIDEBIN
else
    echo "Downloading glide"
    curl https://glide.sh/get | sh
    mkdir -p $GLIDE_CACHE
    cp $GLIDEBIN $GLIDE_CACHE
fi
