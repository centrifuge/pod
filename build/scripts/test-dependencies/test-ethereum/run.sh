#!/usr/bin/env bash

echo "geth node was running? [${GETH_DOCKER_CONTAINER_WAS_RUNNING}]"
if [ -n "${GETH_DOCKER_CONTAINER_WAS_RUNNING}" ]; then
    echo "Container ${GETH_DOCKER_CONTAINER_NAME} is already running. Not starting again."
    exit 0;
else
    echo "Container ${GETH_DOCKER_CONTAINER_NAME} is not currently running. Going to start."
fi

# Setup
local_dir="$(dirname "$0")"
PARENT_DIR=$(pwd)
source "${local_dir}/env_vars.sh"

################## Run GETH #########################
## Ethereum local POA Dev testnet
ETH_DATADIR=$DATA_DIR "${PARENT_DIR}"/build/scripts/docker/run.sh dev

echo "Waiting for GETH to Start Up ..."
maxCount=$(( CENT_ETHEREUM_GETH_START_TIMEOUT / CENT_ETHEREUM_GETH_START_INTERVAL ))
echo "MaxCount: $maxCount"
count=0
while true
do
  mining=$(docker logs geth-node 2>&1 | grep 'mined potential block')
  if [ "$mining" != "" ]; then
    echo "GETH successfully started"
    break
  elif [ $count -ge $maxCount ]; then
    echo "Timeout Starting out GETH"
    exit 1
  fi
  sleep "$CENT_ETHEREUM_GETH_START_INTERVAL";
  ((count++))
done
