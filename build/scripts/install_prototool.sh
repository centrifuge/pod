#!/usr/bin/env sh

PROTOTOOL_BIN=~/bin/$PROTOTOOL_VERSION
mkdir -p ~/bin/$PROTOTOOL_VERSION/

if [ -e "${PROTOTOOL_BIN}/prototool" ]
then
    echo "Found existing prototool in ${PROTOTOOL_BIN}. Not downloading again - just linking."
else
    echo "Downloading prototool"
    curl -sSL https://github.com/uber/prototool/releases/download/v$PROTOTOOL_VERSION/prototool-$(uname -s)-$(uname -m) > $PROTOTOOL_BIN/prototool && chmod +x $PROTOTOOL_BIN/prototool && $PROTOTOOL_BIN/prototool -h
fi
