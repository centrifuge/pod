#!/usr/bin/env sh

local_dir="$(dirname "$0")"

usage() {
  echo "Usage: $0 mode[rinkeby|centapi]"
  exit 1
}

if [ "$#" -lt 1 ]
then
  usage
fi

mode=$1

ETH_DATADIR=${ETH_DATADIR:-${HOME}/Library/Ethereum}
RPC_PORT=${RPC_PORT:-9545}
WS_PORT=${WS_PORT:-9546}
API_PORT=${API_PORT:-8082}
P2P_PORT=${P2P_PORT:-38202}
DEFAULT_DATADIR=/tmp/centrifuge_data_${API_PORT}.leveldb
API_DATADIR=${API_DATADIR:-$DEFAULT_DATADIR}
DEFAULT_CONFIGDIR="${HOME}/centrifuge/cent-api-${API_PORT}"
API_CONFIGDIR=${API_CONFIGDIR:-$DEFAULT_CONFIGDIR}

ADDITIONAL_CMD="${@:2}"

case "$mode" in
  rinkeby)
    ETH_DATADIR=${ETH_DATADIR}/rinkeby RPC_PORT=$RPC_PORT WS_PORT=$WS_PORT RINKEBY=true \
    docker-compose -f $local_dir/docker-compose-geth.yml up > /tmp/geth.log 2>&1 &
  ;;
  centapi)
    CENT_MODE=$CENT_MODE ADDITIONAL_CMD=$ADDITIONAL_CMD API_DATADIR=$API_DATADIR API_CONFIGDIR=$API_CONFIGDIR \
    API_PORT=$API_PORT P2P_PORT=$P2P_PORT \
    docker-compose -f $local_dir/docker-compose-cent-api.yml up > /tmp/cent-api-${API_PORT}.log 2>&1 &
  ;;
  *) usage
esac
