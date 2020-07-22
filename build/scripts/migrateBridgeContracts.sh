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

BRIDGE_DEPLOYMENT_DIR=$PARENT_DIR/build/chainbridge-deploy/cb-sol-cli

BRIDGE_CONTRACTS_DIR=$BRIDGE_DEPLOYMENT_DIR/chainbridge-solidity
if [ ! -e $BRIDGE_CONTRACTS_DIR/build/contracts/Bridge.json ]; then
    echo "$BRIDGE_CONTRACTS_DIR doesn't exist. Probably no migrations run yet. Forcing migrations."
    MIGRATE='true'
fi

if [[ "X${MIGRATE}" == "Xfalse" ]]; then
    echo "not running Asset handler Migrations"
    exit 0
fi

BRIDGE_DEPLOYMENT_DIR=$PARENT_DIR/build/chainbridge-deploy/cb-sol-cli
cd $BRIDGE_DEPLOYMENT_DIR
GIT_COMMIT=v1.0.0 make install
cd $PARENT_DIR

if [ -z ${CENT_ETHEREUM_DAPP_CONTRACTS_DIR} ]; then
    CENT_ETHEREUM_DAPP_CONTRACTS_DIR=${PARENT_DIR}/build
fi

source "${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/env_vars.sh"

cd $BRIDGE_DEPLOYMENT_DIR
bridgeContracts=$(./index.js deploy --gasLimit 7500000 --all --relayerThreshold 1 --relayers $CENT_BRIDGE_RELAYER --privateKey $CENT_ETHEREUM_PRIVATE_KEY  --url=$CENT_ETHEREUM_NODEURL)
bridgeAddr=$(echo -n "$bridgeContracts" | grep "Bridge:" | awk '{print $2}' | tr -d '\n')
erc20Addr=$(echo -n "$bridgeContracts" | grep "Erc20 Handler:" | awk '{print $3}' | tr -d '\n')
erc721Addr=$(echo -n "$bridgeContracts" | grep "Erc721 Handler:" | awk '{print $3}' | tr -d '\n')
genericAddr=$(echo -n "$bridgeContracts" | grep "Generic Handler:" | awk '{print $3}' | tr -d '\n')
echo "${bridgeContracts}"
echo "bridgeAddr $bridgeAddr" > $PARENT_DIR/localAddresses
echo "erc20Addr $erc20Addr" >> $PARENT_DIR/localAddresses
echo "erc721Addr $erc721Addr" >> $PARENT_DIR/localAddresses
echo "genericAddr $genericAddr" >> $PARENT_DIR/localAddresses

# Deploying assetManager
export ETH_RPC_ACCOUNTS=true
export ETH_GAS=$CENT_ETHEREUM_GASLIMIT
export ETH_KEYSTORE="${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/migrateAccount.json"
export ETH_RPC_URL=$CENT_ETHEREUM_NODEURL
export ETH_PASSWORD="/dev/null"
export ETH_FROM="0x89b0a86583c4444acfd71b463e0d3c55ae1412a5"

cd ${CENT_ETHEREUM_DAPP_CONTRACTS_DIR}/ethereum-bridge-contracts
dapp update
dapp build --extract

assetManagerAddr=$(seth send --create out/BridgeAsset.bin 'BridgeAsset(uint8,address)' "10" "$genericAddr")
echo "assetManager $assetManagerAddr" >> $PARENT_DIR/localAddresses

cb-sol-cli --gasLimit 7500000 --gasPrice 10000000000 --url $ETH_RPC_URL --privateKey $CENT_ETHEREUM_PRIVATE_KEY bridge register-generic-resource --bridge $bridgeAddr --handler $genericAddr --targetContract $assetManagerAddr --resourceId 0x0000000000000000000000000000000cb3858f3e48815bfd35c5347aa3b34c01 --deposit 0x00000000 --execute 0x654cf88c

cd $PARENT_DIR
