#!/usr/bin/env bash

set -e

# Allow passing parent directory as a parameter
PARENT_DIR=$1
if [ -z ${PARENT_DIR} ];
then
    PARENT_DIR=`pwd`
    echo "PARENT DIR ${PARENT_DIR}"
fi

# Even if other `env_vars.sh` might hold this variable
# Let's not count on it and be clear instead
if [ -z ${CENT_ETHEREUM_CONTRACTS_DIR} ]; then
    CENT_ETHEREUM_CONTRACTS_DIR=${PARENT_DIR}/vendor/github.com/centrifuge/centrifuge-ethereum-contracts
fi

# Assure that all the dependencies for the contracts are installed
npm install --pwd ${CENT_ETHEREUM_CONTRACTS_DIR} --prefix=${CENT_ETHEREUM_CONTRACTS_DIR}

# `truffle migrate` will fail if not executed in the sub-dir
cd ${CENT_ETHEREUM_CONTRACTS_DIR}

MIGRATE='false'
# Clear up previous build if force build
if [[ "X${FORCE_MIGRATE}" == "Xtrue" ]]; then
  rm -Rf ./build
  MIGRATE='true'
fi


LOCAL_ETH_CONTRACT_ADDRESSES="${CENT_ETHEREUM_CONTRACTS_DIR}/build/contracts/IdentityFactory.json"
if [ ! -e $LOCAL_ETH_CONTRACT_ADDRESSES ]; then
    echo "$LOCAL_ETH_CONTRACT_ADDRESSES doesn't exist. Probably no migrations run yet. Forcing migrations."
    MIGRATE='true'
fi

if [[ "X${MIGRATE}" == "Xtrue" ]]; then
    echo "Running the Solidity contracts migrations for local geth"
    sleep 30 # allow geth block gas limit to increase to more than 7000000
    ${CENT_ETHEREUM_CONTRACTS_DIR}/scripts/migrate.sh localgeth
    if [ $? -ne 0 ]; then
      exit 1
    fi
fi

cd ${PARENT_DIR}

export FORCE_MIGRATE=$MIGRATE

# deploy bridge contracts
./build/scripts/migrateBridgeContracts.sh

identityFactory=$(< $LOCAL_ETH_CONTRACT_ADDRESSES jq -r '.networks."1337".address')
# deploy dapp smartcontracts
IDENTITY_FACTORY=$identityFactory ./build/scripts/migrateDApp.sh
# add bridge balance
./build/scripts/test-dependencies/bridge/add_balance.sh

export MIGRATION_RAN=true
