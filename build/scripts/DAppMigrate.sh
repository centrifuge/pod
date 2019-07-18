#!/usr/bin/env bash

set -e

ANCHOR_ADDR=$1
if [ -z ${ANCHOR_ADDR} ];
then
    echo "${ANCHOR_ADDR} not set"
    exit 1
fi

# Allow passing parent directory as a parameter
PARENT_DIR=$2
if [ -z ${PARENT_DIR} ];
then
    PARENT_DIR=`pwd`
    echo "PARENT DIR ${PARENT_DIR}"
fi

source "${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/env_vars.sh"

if [ -z ${CENT_ETHEREUM_DAPP_CONTRACTS_DIR} ]; then
    CENT_ETHEREUM_DAPP_CONTRACTS_DIR=${PARENT_DIR}/vendor/github.com/centrifuge/privacy-enabled-erc721
fi

cd $CENT_ETHEREUM_DAPP_CONTRACTS_DIR

dapp update
dapp build

export ETH_RPC_ACCOUNTS=true
export ETH_GAS=$CENT_ETHEREUM_GASLIMIT
export ETH_KEYSTORE="${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/migrateAccount.json"
export ETH_RPC_URL=$CENT_ETHEREUM_NODEURL
export ETH_PASSWORD="/dev/null"
export ETH_FROM="0x89b0a86583c4444acfd71b463e0d3c55ae1412a5"

regAddr=$(dapp create "test/TestNFT" "$ANCHOR_ADDR")

echo $regAddr