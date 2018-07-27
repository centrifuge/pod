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
    CENT_ETHEREUM_CONTRACTS_DIR=$GOPATH/src/github.com/CentrifugeInc/centrifuge-ethereum-contracts
fi

# If the contracts dir doesn't exist - it has likely not been "installed" yet
# Installing it in that case so it is usable
if [ ! -d ${CENT_ETHEREUM_CONTRACTS_DIR} ]; then
	echo "Ethereum contracts folder not found at ${CENT_ETHEREUM_CONTRACTS_DIR}. Checking them out."
	# git clone here instead of `go get` as `go get` defaults back to HTTPS which causes issues
    # with certificate-based github authentication
    mkdir -p ${CENT_ETHEREUM_CONTRACTS_DIR}
    git clone git@github.com:CentrifugeInc/centrifuge-ethereum-contracts.git ${CENT_ETHEREUM_CONTRACTS_DIR}

    # Assure that all the dependencies are installed
    npm install --cwd ${CENT_ETHEREUM_CONTRACTS_DIR} --prefix=${CENT_ETHEREUM_CONTRACTS_DIR}

    echo "Due to a fresh checkout of the contracts, requesting a force of the Solidity migrations"
    if [ -z ${FORCE_MIGRATE} ]; then
        FORCE_MIGRATE='true'
    elif [ ${FORCE_MIGRATE} != 'true' ]; then
        echo "Trying to force migrations but variable is already set to [${FORCE_MIGRATE}]. Error out."
        exit -1
    fi
fi

# TODO - ideally we would avoid 'cd-ing' into another directory, but in this case
# `truffle migrate` will fail if not executed in the sub-dir
cd ${CENT_ETHEREUM_CONTRACTS_DIR}
# Clear up previous build
rm -Rf ./build


# TODO move this out into the test dependencies folder instead of doing it here
LOCAL_ETH_CONTRACT_ADDRESSES="${CENT_ETHEREUM_CONTRACTS_DIR}/deployments/local.json"
if [ ! -e $LOCAL_ETH_CONTRACT_ADDRESSES ]; then
    echo "$LOCAL_ETH_CONTRACT_ADDRESSES doesn't exist. Probably no migrations run yet. Forcing migrations."
    FORCE_MIGRATE='true'
fi

if [[ "X${FORCE_MIGRATE}" == "Xtrue" ]];
then
    echo "Running the Solidity contracts migrations for local geth"
    ${CENT_ETHEREUM_CONTRACTS_DIR}/scripts/migrate.sh local
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
    if [ $statusAux -ne 0 ]; then
        echo "Test suite encountered an error. Code [${statusAux}]. Aborting tests."
        break
    fi
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

    # Cleaning extra DAG file, so we do not cache it - travis
    if [[ "X${TRAVIS}" == "Xtrue" ]];
    then
      new_dag=`ls -ltr $DATA_DIR/$NETWORK_ID/.ethash/* | tail -1 | awk '{print $9}' | tr -d '\n'`
      rm -Rf $new_dag
    fi
fi
############################################################

################# Propagate test status ####################
exit $status
############################################################
