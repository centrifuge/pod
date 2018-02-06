#!/usr/bin/env sh
DEP_VERSION="0.4.1"

DEP_REMOTE=https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64

DEPBIN_CACH_FOLDER=$GOPATH/pkg/depbin
DEPBIN_CACHED=$DEPBIN_CACH_FOLDER/dep
DEPBIN=$GOPATH/bin/dep

GO_ETH_CACHED=$GOPATH/src/github.com/ethereum/go-ethereum

if [ ! -e "$DEPBIN_CACHED" ]
then
    echo "Downloading ${DEP_REMOTE} and writing to ${DEPBIN_CACHED}"
    mkdir -p $DEPBIN_CACH_FOLDER
    # Download the binary to bin folder in $GOPATH
    curl -L -s $DEP_REMOTE -o $DEPBIN_CACHED
    # Make the binary executable
    chmod +x $DEPBIN_CACHED
    ln -s $DEPBIN_CACHED $DEPBIN
else
    echo "Found existing dep binary in ${DEPBIN}. Not downloading again - just linking."
    ln -s $DEPBIN_CACHED $DEPBIN
fi

if [ ! -d "$GO_ETH_CACHED" ]
then
    echo "Fetching go-ethereum dependency as not found in ${GO_ETH_CACHED}"
    go get github.com/ethereum/go-ethereum
else
    echo "Found ${GO_ETH_CACHED}. Not fetching again."
fi

