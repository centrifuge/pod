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

if [ ! -e $PARENT_DIR/localAddresses ]; then
    echo "$PARENT_DIR/localAddresses doesn't exist. Probably no migrations run yet. Forcing migrations."
    MIGRATE='true'
fi

if [[ "X${MIGRATE}" == "Xfalse" ]]; then
    echo "not running Dapp Migrations"
    exit 0
fi

source "${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/env_vars.sh"

if [ -z ${CENT_ETHEREUM_DAPP_CONTRACTS_DIR} ]; then
    CENT_ETHEREUM_DAPP_CONTRACTS_DIR=${PARENT_DIR}/build
fi

#ASSET_DIR=${CENT_ETHEREUM_DAPP_CONTRACTS_DIR}/ethereum-bridge-contracts
NFT_DIR=${CENT_ETHEREUM_DAPP_CONTRACTS_DIR}/privacy-enabled-erc721

export ETH_RPC_ACCOUNTS=true
export ETH_GAS=$CENT_ETHEREUM_GASLIMIT
export ETH_KEYSTORE="${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/migrateAccount.json"
export ETH_RPC_URL=$CENT_ETHEREUM_NODEURL
export ETH_PASSWORD="/dev/null"
export ETH_FROM="0x89b0a86583c4444acfd71b463e0d3c55ae1412a5"

## deploy asset contracts
#cd $ASSET_DIR
#dapp update
#dapp build --extract
#
#assetAddr=$(seth send --create out/BridgeAsset.bin 'BridgeAsset(uint8)' "10")

# deploy NFT contract
cd $NFT_DIR
dapp update
dapp build --extract

echo "Identity factory $IDENTITY_FACTORY"
assetManagerAddr=$(cat $PARENT_DIR/localAddresses | grep "assetManager" | awk '{print $2}' | tr -d '\n')
nftAddr=$(seth send --create out/AssetNFT.bin 'AssetNFT(address, address)' "$assetManagerAddr" "$IDENTITY_FACTORY")
#echo "assetManager $assetAddr" > $PARENT_DIR/localAddresses
echo "genericNFT $nftAddr" >> $PARENT_DIR/localAddresses

bridgeAddr=$(cat $PARENT_DIR/localAddresses | grep "bridgeAddr" | awk '{print $2}' | tr -d '\n')
erc721Addr=$(cat $PARENT_DIR/localAddresses | grep "erc721Addr" | awk '{print $2}' | tr -d '\n')
erc20Addr=$(cat $PARENT_DIR/localAddresses | grep "erc20Addr" | awk '{print $2}' | tr -d '\n')
echo "creating bridge config with bridge addresses $bridgeAddr $erc721Addr $erc20Addr $assetManagerAddr"
bridge_dir="$PARENT_DIR"/build/scripts/test-dependencies/bridge
"$bridge_dir"/create_config.sh "$bridge_dir"/config "$bridgeAddr" "$erc721Addr" "$erc20Addr" "$assetManagerAddr"

cd $PARENT_DIR
