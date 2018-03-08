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
geth --identity "${IDENTITY}" --nodiscover --networkid=$NETWORK_ID --datadir=${DATA_DIR} init ${PARENT_DIR}/scripts/test-dependencies/test-ethereum/genesis.json
cp ${PARENT_DIR}/scripts/test-dependencies/test-ethereum/coinbase.json $DATA_DIR/keystore/
cp ${PARENT_DIR}/scripts/test-dependencies/test-ethereum/userAccount.json $DATA_DIR/keystore/
#
################## Run GETH #########################
## Ethereum local testnet
geth --identity "${IDENTITY}" --nodiscover --networkid=$NETWORK_ID --datadir=${DATA_DIR} --cache=512 --rpc --rpcport $RPC_PORT --rpcapi="db,eth,net,personal,web3" --mine --etherbase "${CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS}" &> $DATA_DIR/geth.out &

# Wait until DAG has been generated
maxCount=300 # Wait 10 minutes max
count=0
while true
do
  finished=`cat $DATA_DIR/geth.out | grep 'Generated ethash verification cache'`
  mining=`cat $DATA_DIR/geth.out | grep 'mined potential block'`
  if [ "$finished" != "" ] || [ "$mining" != "" ]; then
    echo "GETH successfully started"
    break
  elif [ $count -ge $maxCount ]; then
    echo "Timeout Starting out GETH"
    exit 1
  fi
  sleep 2;
  ((count++))
done
