#!/bin/bash

source ./env_vars.sh

mkdir -p $DATA_DIR

geth --identity "${IDENTITY}" --nodiscover --networkid=$NETWORK_ID --datadir=$DATA_DIR  init genesis.json
