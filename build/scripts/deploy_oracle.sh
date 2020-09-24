#!/usr/bin/env bash

set -e

# Allow passing parent directory as a parameter
PARENT_DIR=$1
if [ -z ${PARENT_DIR} ];
then
    PARENT_DIR=`pwd`
fi

if [ -z ${CENT_ETHEREUM_DAPP_CONTRACTS_DIR} ]; then
    CENT_ETHEREUM_DAPP_CONTRACTS_DIR=${PARENT_DIR}/build
fi

source "${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/env_vars.sh"
ORACLE_DIR=${CENT_ETHEREUM_DAPP_CONTRACTS_DIR}/chainlink-oracle-contract

export ETH_RPC_ACCOUNTS=true
export ETH_GAS=$CENT_ETHEREUM_GASLIMIT
export ETH_KEYSTORE="${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/migrateAccount.json"
export ETH_RPC_URL=$CENT_ETHEREUM_NODEURL
export ETH_PASSWORD="/dev/null"
export ETH_FROM="0x89b0a86583c4444acfd71b463e0d3c55ae1412a5"

# deploy NFT contract
cd "$ORACLE_DIR"
dapp update
dapp --use solc:0.5.0 build --extract

# deploy mock contract if not set
if [[ -z "${NFT_UPDATE}" ]]; then
  echo "Deploying NFTUpdate contract"
  NFT_UPDATE=$(seth send --create out/NFTUpdate.bin 'NFTUpdate()')
  set +ex
  echo "-------------------------------------------------------------------------------"
  echo "NFTUpdate deployed to: $NFT_UPDATE"
  echo "-------------------------------------------------------------------------------"
fi

WARDS=[]
REGISTRY=$(cat $PARENT_DIR/localAddresses | grep "genericNFT" | awk '{print $2}' | tr -d '\n')
ORACLE=$(seth send --create out/NFTOracle.bin 'NFTOracle(address,address,bytes32,address[])' $NFT_UPDATE $REGISTRY $FINGERPRINT $WARDS)

seth send $ORACLE 'rely(address)' $OWNER

echo "oracle $ORACLE" >> $PARENT_DIR/localAddresses

cd $PARENT_DIR
