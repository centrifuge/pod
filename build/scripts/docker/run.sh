#!/usr/bin/env bash

local_dir="$(dirname "$0")"

usage() {
  echo "Usage: $0 mode[init|rinkeby|local|mine|centapi]"
  exit 1
}

if [ "$#" -lt 1 ]
then
  usage
fi

mode=$1

ETH_DATADIR=${ETH_DATADIR:-${HOME}/Library/Ethereum}
BOOT_NODES=${BOOT_NODES:-'enode://2597b631806959d5d40aa07797754338ee98809c51de58024334b5951114fb013cf1a527ee941b80d776c2cdf257e9c39d31414c0efd0051f75a4c9afdcd8e9c@35.192.161.113:30303'}
NETWORK_ID=${NETWORK_ID:-1337}
IDENTITY=${IDENTITY:-CentTestEth}
API=${API:-'db,eth,net,web3,personal,txpool'}
RPC_PORT=${RPC_PORT:-9545}
WS_PORT=${WS_PORT:-9546}
CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS=${CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS:-'0x89b0a86583c4444acfd71b463e0d3c55ae1412a5'}
API_PORT=${API_PORT:-8082}
P2P_PORT=${P2P_PORT:-38202}
DEFAULT_DATADIR=/tmp/centrifuge_data_${API_PORT}.leveldb
API_DATADIR=${API_DATADIR:-$DEFAULT_DATADIR}
DEFAULT_CONFIGDIR="${HOME}/centrifuge/cent-api-${API_PORT}"
API_CONFIGDIR=${API_CONFIGDIR:-$DEFAULT_CONFIGDIR}
TARGETGASLIMIT=${TARGETGASLIMIT:-"9000000"}

ADDITIONAL_CMD="${@:2}"

case "$mode" in
  init)
    mkdir -p ${ETH_DATADIR}/${NETWORK_ID}/files
    mkdir -p ${ETH_DATADIR}/${NETWORK_ID}/keystore
    mkdir -p ${ETH_DATADIR}/${NETWORK_ID}/.ethash
    if [ ! -f ${ETH_DATADIR}/${NETWORK_ID}/files/genesis.json ]; then
      cp $local_dir/../test-dependencies/test-ethereum/genesis.json ${ETH_DATADIR}/${NETWORK_ID}/files
    fi

    INIT_ETH=true IDENTITY=$IDENTITY NETWORK_ID=$NETWORK_ID ETH_DATADIR=${ETH_DATADIR}/${NETWORK_ID} \
    docker-compose -f $local_dir/docker-compose-geth-init.yml up
  ;;
  local)
    IDENTITY=$IDENTITY NETWORK_ID=$NETWORK_ID ETH_DATADIR=${ETH_DATADIR}/${NETWORK_ID} GETH_LOCAL=true RPC_PORT=$RPC_PORT WS_PORT=$WS_PORT \
    BOOT_NODES=$BOOT_NODES \
    docker-compose -f $local_dir/docker-compose-geth.yml up > /tmp/geth.log 2>&1 &
  ;;
  mine)
    cp $local_dir/../test-dependencies/test-ethereum/*.json ${ETH_DATADIR}/${NETWORK_ID}/keystore

    IDENTITY=$IDENTITY NETWORK_ID=$NETWORK_ID ETH_DATADIR=${ETH_DATADIR}/${NETWORK_ID} RPC_PORT=$RPC_PORT API=$API \
    WS_PORT=$WS_PORT CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS=$CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS GETH_MINE=true \
    docker-compose -f $local_dir/docker-compose-geth.yml up > /tmp/geth.log 2>&1 &
  ;;
  dev)
    mkdir -p ${ETH_DATADIR}/${NETWORK_ID}/keystore
    cp $local_dir/../test-dependencies/test-ethereum/*.json ${ETH_DATADIR}/${NETWORK_ID}/keystore
    IDENTITY=$IDENTITY NETWORK_ID=$NETWORK_ID ETH_DATADIR=${ETH_DATADIR}/${NETWORK_ID} RPC_PORT=$RPC_PORT API=$API \
    WS_PORT=$WS_PORT CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS=$CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS \
    TARGETGASLIMIT=$TARGETGASLIMIT docker-compose -f $local_dir/docker-compose-geth.yml up > /tmp/geth.log 2>&1 &
  ;;
  centapi)
    CENT_MODE=$CENT_MODE ADDITIONAL_CMD=$ADDITIONAL_CMD API_DATADIR=$API_DATADIR API_CONFIGDIR=$API_CONFIGDIR \
    API_PORT=$API_PORT P2P_PORT=$P2P_PORT \
    docker-compose -f $local_dir/docker-compose-cent-api.yml up > /tmp/cent-api-${API_PORT}.log 2>&1 &
  ;;
  ccdev)
    docker-compose -f $local_dir/docker-compose-cc.yml up > /tmp/cc-0.log 2>&1 &
  ;;
  bridge)
    BRIDGE_CONFIGDIR=$local_dir/../test-dependencies/bridge/config/
    BRIDGE_KEYSDIR=$local_dir/../test-dependencies/bridge/keys/
    BRIDGE_CONFIGDIR=$BRIDGE_CONFIGDIR BRIDGE_KEYSDIR=$BRIDGE_KEYSDIR docker-compose -f $local_dir/docker-compose-bridge.yml up > /tmp/bridge-0.log 2>&1 &
  ;;
  *) usage
esac
echo "Done"
