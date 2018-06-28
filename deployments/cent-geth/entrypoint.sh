#!/usr/bin/env sh

set -x

API=${API:-"db,eth,net,web3,txpool"}

if [[ "X$INIT_ETH" = "Xtrue" ]];
then
  /geth --identity $IDENTITY --datadir /root/.ethereum --ethash.dagdir /root/.ethereum/.ethash --networkid $NETWORK_ID init /root/.ethereum/files/genesis.json
  mkdir -p /root/.ethereum/keystore
  if [ -d /root/.ethereum/secrets ]; then
    cp /root/.ethereum/secrets/*.json /root/.ethereum/keystore
  fi
elif [[ "X$RINKEBY" = "Xtrue" ]];
then
  /geth --rinkeby --light --rpc --rpcport $RPC_PORT --rpcaddr 0.0.0.0 --rpcapi $API \
        --ws --wsport $WS_PORT --wsaddr 0.0.0.0 --wsorigins "*" --wsapi $API \
        --datadir /root/.ethereum --ethash.dagdir /root/.ethereum/.ethash
elif [[ "X$GETH_LOCAL" = "Xtrue" ]];
then
  /geth --identity "${IDENTITY}" --networkid $NETWORK_ID --rpc --rpcport $RPC_PORT --rpcaddr 0.0.0.0 --rpcapi db,eth,net,web3,personal,admin,txpool \
      --ws --wsport $WS_PORT --wsaddr 0.0.0.0 --wsorigins "*" --wsapi db,eth,net,web3,personal,admin,txpool \
      --datadir /root/.ethereum --ethash.dagdir /root/.ethereum/.ethash --cache 512 \
      --bootnodes "${BOOT_NODES}"
else
  /geth --identity "${IDENTITY}" --networkid $NETWORK_ID --datadir /root/.ethereum --ethash.dagdir /root/.ethereum/.ethash \
        --cache 512 --rpc --rpcaddr 0.0.0.0 --rpcport $RPC_PORT --rpcapi $API --mine --etherbase "${CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS}" \
        --ipcdisable --ws --wsport $WS_PORT --wsaddr 0.0.0.0 --wsorigins "*" --wsapi $API --gasprice "40000"
fi