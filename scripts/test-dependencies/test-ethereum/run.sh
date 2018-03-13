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

####### Ensure Geth Docker Image is present #########
exists=`docker images ethereum/client-go:$GETH_DOCKER_VERSION -q`
if [ "X${exists}" != "X" ]
then
  echo "Found existing docker image [ethereum/client-go:$GETH_DOCKER_VERSION] Not downloading again."
else
  docker pull ethereum/client-go:$GETH_DOCKER_VERSION
fi
#####################################################

################# Init GETH #########################
DOCKER_DATA_DIR=/root/.ethereum
DOCKER_DAG_DIR=/root/.ethash
mkdir -p $DATA_DIR/keystore
cp ${PARENT_DIR}/scripts/test-dependencies/test-ethereum/coinbase.json $DATA_DIR/keystore/
cp ${PARENT_DIR}/scripts/test-dependencies/test-ethereum/userAccount.json $DATA_DIR/keystore/
cp ${PARENT_DIR}/scripts/test-dependencies/test-ethereum/migrateAccount.json $DATA_DIR/keystore/
docker run -it -v $DATA_DIR:/$DOCKER_DATA_DIR -v $PARENT_DIR/scripts/test-dependencies/test-ethereum:$DOCKER_DATA_DIR/files ethereum/client-go:$GETH_DOCKER_VERSION --identity "${IDENTITY}" --nodiscover --networkid=$NETWORK_ID --datadir=${DOCKER_DATA_DIR} init ${DOCKER_DATA_DIR}/files/genesis.json

################## Run GETH #########################
## Ethereum local testnet
docker run -d --name geth-node -p 9545:9545 -p 9546:9546 -p 30303:30303 -v $DATA_DIR:/$DOCKER_DATA_DIR -v $HOME/.ethash:$DOCKER_DAG_DIR ethereum/client-go:$GETH_DOCKER_VERSION --identity "${IDENTITY}" --nodiscover --networkid=$NETWORK_ID --datadir=${DOCKER_DATA_DIR} --cache=512 --rpc --rpcaddr 0.0.0.0 --rpcport $RPC_PORT --rpcapi="db,eth,net,personal,web3" --mine --etherbase "${CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS}" --ipcdisable --ws --wsport $WS_PORT --wsaddr 0.0.0.0 --wsorigins "*" --wsapi="db,eth,net,personal,web3"

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
