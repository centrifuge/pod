#!/usr/bin/env bash

# Allow passing parent directory as a parameter
PARENT_DIR=$1
if [ -z ${PARENT_DIR} ];
then
    echo "PARENT DIR $1"
    PARENT_DIR = `pwd`
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
# Clear up previous build
rm -Rf ./build


LOCAL_ETH_CONTRACT_ADDRESSES="${CENT_ETHEREUM_CONTRACTS_DIR}/deployments/local.json"
if [ ! -e $LOCAL_ETH_CONTRACT_ADDRESSES ]; then
    echo "$LOCAL_ETH_CONTRACT_ADDRESSES doesn't exist. Probably no migrations run yet. Forcing migrations."
    FORCE_MIGRATE='true'
fi

if [[ "X${FORCE_MIGRATE}" == "Xtrue" ]];
then
    echo "Running the Solidity contracts migrations for local geth"
    ${CENT_ETHEREUM_CONTRACTS_DIR}/scripts/migrate.sh localgeth
else
    echo "Not migrating the Solidity contracts"
fi

cd ${PARENT_DIR}