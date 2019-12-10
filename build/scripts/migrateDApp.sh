#!/usr/bin/env bash

set -e

# Allow passing parent directory as a parameter
PARENT_DIR=$1
if [ -z ${PARENT_DIR} ];
then
    PARENT_DIR=`pwd`
fi

source "${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/env_vars.sh"

if [ -z ${CENT_ETHEREUM_DAPP_CONTRACTS_DIR} ]; then
    CENT_ETHEREUM_DAPP_CONTRACTS_DIR=${PARENT_DIR}/build/privacy-enabled-erc721
fi

cd $CENT_ETHEREUM_DAPP_CONTRACTS_DIR

dapp update
dapp build --extract

export ETH_RPC_ACCOUNTS=true
export ETH_GAS=$CENT_ETHEREUM_GASLIMIT
export ETH_KEYSTORE="${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/migrateAccount.json"
export ETH_RPC_URL=$CENT_ETHEREUM_NODEURL
export ETH_PASSWORD="/dev/null"
export ETH_FROM="0x89b0a86583c4444acfd71b463e0d3c55ae1412a5"

regAddr=$(seth send --create out/NFT.bin 'NFT(string memory, string memory)' "CentNFT" "CentNFT")

echo -n "genericNFT $regAddr" > $PARENT_DIR/localAddresses

cd $PARENT_DIR
