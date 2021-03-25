#!/usr/bin/env bash

set -e

# Allow passing parent directory as a parameter
PARENT_DIR=$1
if [ -z "${PARENT_DIR}" ];
then
    PARENT_DIR=$(pwd)
fi

MIGRATE='false'

# Clear up previous build if force build
if [[ "${FORCE_MIGRATE}" == "true" ]]; then
  MIGRATE='true'
fi

if [ ! -e "$PARENT_DIR"/localAddresses ]; then
    echo "$PARENT_DIR/localAddresses doesn't exist. Probably no migrations run yet. Forcing migrations."
    MIGRATE='true'
fi

if [[ "${MIGRATE}" == "false" ]]; then
    echo "not running Dapp Migrations"
    exit 0
fi

source "${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/env_vars.sh"

if [ -z "${CENT_ETHEREUM_DAPP_CONTRACTS_DIR}" ]; then
    CENT_ETHEREUM_DAPP_CONTRACTS_DIR=${PARENT_DIR}/build
fi

NFT_DIR=${CENT_ETHEREUM_DAPP_CONTRACTS_DIR}/privacy-enabled-erc721

export ETH_RPC_ACCOUNTS=true
export ETH_GAS=$CENT_ETHEREUM_GASLIMIT
export ETH_KEYSTORE="${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/migrateAccount.json"
export ETH_RPC_URL=$CENT_ETHEREUM_NODEURL
export ETH_PASSWORD="/dev/null"
export ETH_FROM="0x89b0a86583c4444acfd71b463e0d3c55ae1412a5"

# deploy NFT contract
cd "$NFT_DIR"
dapp --use solc:0.5.15 update
dapp --use solc:0.5.15 build --extract

echo "Identity factory $IDENTITY_FACTORY"
assetManagerAddr=$(< "$PARENT_DIR"/localAddresses grep "assetManager" | awk '{print $2}' | tr -d '\n')
nftAddr=$(seth send --create out/AssetNFT.bin 'AssetNFT(address, address)' "$assetManagerAddr" "$IDENTITY_FACTORY")
echo "genericNFT $nftAddr" >> "$PARENT_DIR"/localAddresses

genericAddr=$(< "$PARENT_DIR"/localAddresses grep "genericAddr" | awk '{print $2}' | tr -d '\n')
bridgeAddr=$(< "$PARENT_DIR"/localAddresses grep "bridgeAddr" | awk '{print $2}' | tr -d '\n')
erc721Addr=$(< "$PARENT_DIR"/localAddresses grep "erc721Addr" | awk '{print $2}' | tr -d '\n')
erc20Addr=$(< "$PARENT_DIR"/localAddresses grep "erc20Addr" | awk '{print $2}' | tr -d '\n')
echo "creating bridge config with bridge addresses $bridgeAddr $erc721Addr $erc20Addr $genericAddr"
bridge_dir="$PARENT_DIR"/build/scripts/test-dependencies/bridge
"$bridge_dir"/create_config.sh "$bridge_dir"/config "$bridgeAddr" "$erc721Addr" "$erc20Addr" "$genericAddr"

cd "$PARENT_DIR"
