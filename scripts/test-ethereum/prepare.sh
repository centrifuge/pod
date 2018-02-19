#!/bin/bash

source ./env_vars.sh

mkdir -p $DATA_DIR
mkdir -p $DATA_DIR/keystore

geth --identity "${IDENTITY}" --nodiscover --networkid=$NETWORK_ID --datadir=$DATA_DIR  init genesis.json

#password := `ZhXfpAc#vHu4JTELA`
cp coinbase.json $DATA_DIR/keystore/

#password := `fenrwf34nr3cdlsmk`
cp userAccount.json $DATA_DIR/keystore/