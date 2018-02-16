#!/bin/bash

source env_vars.sh

geth --identity "${IDENTITY}" --nodiscover --networkid=$NETWORK_ID --datadir=$DATA_DIR --cache=512 --rpc --rpcport $RPC_PORT --rpcapi="db,eth,net,personal,web3" --mine --etherbase '838f7dca284eb69a9c489fe09c31cff37defdeca'
