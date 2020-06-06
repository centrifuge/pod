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

BRIDGE_CONTRACTS_DIR=$PARENT_DIR/build/chainbridge-solidity
if [ ! -e $BRIDGE_CONTRACTS_DIR/build/contracts/Bridge.json ]; then
    echo "$BRIDGE_CONTRACTS_DIR doesn't exist. Probably no migrations run yet. Forcing migrations."
    MIGRATE='true'
fi

if [[ "X${MIGRATE}" == "Xfalse" ]]; then
    echo "not running Asset handler Migrations"
    exit 0
fi

if [ -z ${CENT_ETHEREUM_DAPP_CONTRACTS_DIR} ]; then
    CENT_ETHEREUM_DAPP_CONTRACTS_DIR=${PARENT_DIR}/build
fi

source "${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/env_vars.sh"

cd $BRIDGE_CONTRACTS_DIR
make install-deps
make install-cli
make compile
bridgeContracts=$(./cli/index.js deploy --relayer-threshold 1 --relayers $CENT_BRIDGE_RELAYER --private-key $CENT_ETHEREUM_PRIVATE_KEY  --url=$CENT_ETHEREUM_NODEURL)
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

# This code will be changed to deploy the assetHash contract only passing relayers check of generic handler
abi=$(jq -r ".abi" $PARENT_DIR/build/chainbridge-solidity/build/contracts/Bridge.json)
cat >$PARENT_DIR/build/scripts/initBridge.js << EOF
var abi = $abi ;
var cc = web3.eth.contract(abi).at("$bridgeAddr");
cc.adminSetGenericResource("$genericAddr", "0x0000000000000000000000000000000cb3858f3e48815bfd35c5347aa3b34c01", "$assetManagerAddr", "0x00", "0x654cf88c", {gas: 1000000, from: "0x89b0a86583c4444acfd71b463e0d3c55ae1412a5"});
EOF

cat $PARENT_DIR/build/scripts/initBridge.js
docker run --net=host --entrypoint "/geth" centrifugeio/cent-geth:v0.1.1 attach http://localhost:9545 --exec "personal.unlockAccount('0x89b0a86583c4444acfd71b463e0d3c55ae1412a5', '${MIGRATE_PASSWORD}', 500)"
docker run --net=host --entrypoint "/geth" -v $PARENT_DIR/build/scripts:/tmp centrifugeio/cent-geth:v0.1.1 attach http://localhost:9545 --jspath "/tmp" --exec 'loadScript("initBridge.js")'

rm $PARENT_DIR/build/scripts/initBridge.js

cd $PARENT_DIR
