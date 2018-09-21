#!/usr/bin/env bash

set -a

## run this so that latest local code is rebuilt
make install

# Set local contracts directory
export CENT_ETHEREUM_CONTRACTS_DIR=${PARENT_DIR}/vendor/github.com/centrifuge/centrifuge-ethereum-contracts

################# Prepare for run ########################
PARENT_DIR=`pwd`
source "${PARENT_DIR}/scripts/setup_smart_contract_addresses.sh"

centrifuge run --config example/resources/centrifuge_example.yaml
