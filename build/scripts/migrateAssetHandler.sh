#!/usr/bin/env bash

set -e

# Allow passing parent directory as a parameter
PARENT_DIR=$1
if [ -z ${PARENT_DIR} ];
then
    PARENT_DIR=`pwd`
fi

MIGRATE='false'

# Clear up previous build if force build
if [[ "X${FORCE_MIGRATE}" == "Xtrue" ]]; then
  MIGRATE='true'
fi

CENTRIFUGE_HANDLER_DIR=$PARENT_DIR/build/chainbridge-solidity
if [ ! -e $CENTRIFUGE_HANDLER_DIR/build/contracts/CentrifugeAssetHandler.json ]; then
    echo "$CENTRIFUGE_HANDLER_DIR doesn't exist. Probably no migrations run yet. Forcing migrations."
    MIGRATE='true'
fi

if [[ "X${MIGRATE}" == "Xfalse" ]]; then
    echo "not running Asset handler Migrations"
    exit 0
fi

source "${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/env_vars.sh"
cd $CENTRIFUGE_HANDLER_DIR
make install-deps
make install-cli
make compile
hanlderContract=$(./cli/index.js deploy --private-key $CENT_ETHEREUM_PRIVATE_KEY  --url=$CENT_ETHEREUM_NODEURL | grep "Centrifuge Handler: " | awk '{print $3}')
echo -n "assetHandler $hanlderContract" >> $PARENT_DIR/localAddresses
cd $PARENT_DIR
