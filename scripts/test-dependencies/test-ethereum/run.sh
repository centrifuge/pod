#!/bin/bash

# Setup
local_dir="$(dirname "$0")"
PARENT_DIR=`pwd`
source "${local_dir}/env_vars.sh"

# For caching when running from travis
if [[ "X${RUN_CONTEXT}" == "Xtravis" ]];
then
  ln -s $DATA_DIR/.ethash ~/.ethash
  echo "TRAVIS"
else
  echo "LOCAL CONTEXT"
fi

################# Init GETH #########################
ETH_DATADIR=$DATA_DIR ${PARENT_DIR}/scripts/docker/run.sh init

################## Run GETH #########################
## Ethereum local testnet
ETH_DATADIR=$DATA_DIR API='db,eth,net,web3,personal' ${PARENT_DIR}/scripts/docker/run.sh mine

echo "Waiting for GETH to Start Up ..."
# Wait until DAG has been generated
maxCount=$(( CENT_ETHEREUM_GETH_START_TIMEOUT / $CENT_ETHEREUM_GETH_START_INTERVAL ))
echo "MaxCount: $maxCount"
count=0
while true
do
  finished=`docker logs geth-node 2>&1 | grep 'Generated ethash verification cache'`
  mining=`docker logs geth-node 2>&1 | grep 'mined potential block'`
  if [ "$finished" != "" ] || [ "$mining" != "" ]; then
    echo "GETH successfully started"
    break
  elif [ $count -ge $maxCount ]; then
    echo "Timeout Starting out GETH"
    exit 1
  fi
  sleep $CENT_ETHEREUM_GETH_START_INTERVAL;
  ((count++))
done
