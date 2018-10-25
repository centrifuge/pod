#!/usr/bin/env bash
set -a

# Setup
local_dir="$(dirname "$0")"
PARENT_DIR=`pwd`

if [[ "X${1}" == "Xmigrate" ]] || [[ "X${TRAVIS}" == "Xtrue" ]];
then
  FORCE_MIGRATE='true'
fi

GETH_DOCKER_CONTAINER_NAME="geth-node"
GETH_DOCKER_CONTAINER_WAS_RUNNING=`docker ps -a --filter "name=${GETH_DOCKER_CONTAINER_NAME}" --filter "status=running" --quiet`

# Code coverage is stored in coverage.txt
echo "" > coverage.txt

################# Run Dependencies #########################
for path in ${local_dir}/test-dependencies/*; do
    [ -d "${path}" ] || continue # if not a directory, skip
    source "${path}/env_vars.sh" # Every dependency should have env_vars.sh + run.sh executable files
    echo "Executing [${path}/run.sh]"
    ${path}/run.sh
    if [ $? -ne 0 ]; then
        exit 1
    fi
done
############################################################

################# Prepare for tests ########################
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
status=$?

cd ${PARENT_DIR}

############################################################

################# Run Tests ################################
if [ $status -eq 0 ]; then
  statusAux=0
  for path in ${local_dir}/tests/*; do
    [ -x "${path}" ] || continue # if not an executable, skip

    echo "Executing test suite [${path}]"
    ./$path
    statusAux="$(( $statusAux | $? ))"
  done
  # Store status of tests
  status=$statusAux
fi
############################################################

################# CleanUp ##################################
if [ -n "${GETH_DOCKER_CONTAINER_WAS_RUNNING}" ]; then
    echo "Container ${GETH_DOCKER_CONTAINER_NAME} was already running before the test setup. Not tearing it down as the assumption is that the container was started outside this context."
else
    echo "Bringing GETH Daemon Down"
    docker rm -f geth-node
fi
############################################################

################# Propagate test status ####################
echo "The test suite overall is exiting with status [$status]"
exit $status
############################################################
