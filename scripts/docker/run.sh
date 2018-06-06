#!/usr/bin/env sh

local_dir="$(dirname "$0")"

usage() {
  echo "Usage: $0 mode[init|rinkeby|local|mine]"
  exit 1
}

if [ "$#" -ne 1 ]
then
  usage
fi

mode=$1

ETH_DATADIR=${ETH_DATADIR:-${HOME}/Library/Ethereum}
BOOT_NODES=${BOOT_NODES:-'enode://0b6892ea2bc5269e1f92333bab695ad3df10a48693bb789bd096d46247957ae4f7ca06234e973c92b39654f646fff278756f31dc0ffe2fdfe6e769d77bf1bafd@35.192.161.113:30303'}
NETWORK_ID=${NETWORK_ID:-8383}
IDENTITY=${IDENTITY:-CentTestEth}
API=${API:-'db,eth,net,web3'}
RPC_PORT=${RPC_PORT:-9545}
WS_PORT=${WS_PORT:-9546}
CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS=${CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS:-'0x4b1b843af77a8f7f4f0ad2145095937e3d90e3d8'}

case "$mode" in
  init)
    mkdir -p ${ETH_DATADIR}/${NETWORK_ID}/files
    mkdir -p ${ETH_DATADIR}/${NETWORK_ID}/keystore
    mkdir -p ${ETH_DATADIR}/${NETWORK_ID}/.ethash
    if [ ! -f ${ETH_DATADIR}/${NETWORK_ID}/files/genesis.json ]; then
      cp $local_dir/../test-dependencies/test-ethereum/genesis.json ${ETH_DATADIR}/${NETWORK_ID}/files
    fi

    INIT_ETH=true IDENTITY=$IDENTITY NETWORK_ID=$NETWORK_ID ETH_DATADIR=${ETH_DATADIR}/${NETWORK_ID} \
    docker-compose -f $local_dir/docker-compose-init.yml up
  ;;
  rinkeby)
    ETH_DATADIR=${ETH_DATADIR}/rinkeby RPC_PORT=$RPC_PORT RINKEBY=true \
    docker-compose -f $local_dir/docker-compose.yml up > /tmp/geth.log 2>&1 &
  ;;
  local)
    IDENTITY=$IDENTITY NETWORK_ID=$NETWORK_ID ETH_DATADIR=${ETH_DATADIR}/${NETWORK_ID} GETH_LOCAL=true RPC_PORT=$RPC_PORT \
    BOOT_NODES=$BOOT_NODES \
    docker-compose -f $local_dir/docker-compose.yml up > /tmp/geth.log 2>&1 &
  ;;
  mine)
    ls -la ${ETH_DATADIR}/${NETWORK_ID}
    whoami
    cp $local_dir/../test-dependencies/test-ethereum/*.json ${ETH_DATADIR}/${NETWORK_ID}/keystore

    IDENTITY=$IDENTITY NETWORK_ID=$NETWORK_ID ETH_DATADIR=${ETH_DATADIR}/${NETWORK_ID} RPC_PORT=$RPC_PORT API=$API \
    WS_PORT=$WS_PORT CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS=$CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS \
    docker-compose -f $local_dir/docker-compose.yml up > /tmp/geth.log 2>&1 &
  ;;
  *) usage
esac
