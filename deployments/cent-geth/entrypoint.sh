#!/usr/bin/env sh

set -x

if [[ "X$INIT_ETH" = "Xtrue" ]];
then
  /geth --identity $IDENTITY --datadir $DOCKER_DATA_DIR --ethash.dagdir $DOCKER_ETHASH_DIR --nodiscover --networkid $NETWORK_ID init $DOCKER_DATA_DIR/files/genesis.json
  mkdir -p $DOCKER_DATA_DIR/keystore
  cp $DOCKER_DATA_DIR/secrets/*.json $DOCKER_DATA_DIR/keystore
else
  /geth --identity "${IDENTITY}" --nodiscover --networkid $NETWORK_ID --datadir ${DOCKER_DATA_DIR} --ethash.dagdir $DOCKER_ETHASH_DIR --cache 512 --rpc --rpcaddr 0.0.0.0 --rpcport $RPC_PORT --rpcapi "db,eth,net,personal,web3" --mine --etherbase "${CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS}" --ipcdisable --ws --wsport $WS_PORT --wsaddr 0.0.0.0 --wsorigins "*" --wsapi "db,eth,net,personal,web3"
fi