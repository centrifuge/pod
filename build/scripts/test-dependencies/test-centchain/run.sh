#!/usr/bin/env bash

echo "centchain node was running? [${CC_DOCKER_CONTAINER_WAS_RUNNING}]"
if [ -n "${CC_DOCKER_CONTAINER_WAS_RUNNING}" ]; then
    echo "Container ${CC_DOCKER_CONTAINER_NAME} is already running. Not starting again."
    exit 0;
else
    echo "Container ${CC_DOCKER_CONTAINER_NAME} is not currently running. Going to start."
fi

# Setup
PARENT_DIR=`pwd`
local_dir="$(dirname "$0")"
source "${local_dir}/env_vars.sh"

################## Run CentChain #########################
## Centrifuge Chain local POA Dev testnet
${PARENT_DIR}/build/scripts/docker/run.sh ccdev

echo "Waiting for Centrifuge Chain to Start Up ..."
maxCount=$(( CENT_ETHEREUM_GETH_START_TIMEOUT / $CENT_ETHEREUM_GETH_START_INTERVAL ))
echo "MaxCount: $maxCount"
count=0
while true
do
  mining=`docker logs cc-node 2>&1 | grep 'finalized #'`
  if [ "$mining" != "" ]; then
    echo "CentChain successfully started"
    break
  elif [ $count -ge $maxCount ]; then
    echo "Timeout Starting out CentChain"
    exit 1
  fi
  sleep $CENT_ETHEREUM_GETH_START_INTERVAL;
  ((count++))
done
