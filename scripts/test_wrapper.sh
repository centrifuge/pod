#!/usr/bin/env bash
set -a

# Setup
local_dir="$(dirname "$0")"
PARENT_DIR=`pwd`
GETH_DOCKER_CONTAINER_NAME="geth-node"
CONTAINER_WAS_RUNNING=`docker ps -a --filter "name=${DOCKER_CONTAINER_NAME}" --filter "status=running" --quiet`


################# Run Dependencies #########################
for path in ${local_dir}/test-dependencies/*; do
    [ -d "${path}" ] || continue # if not a directory, skip
    source "${path}/env_vars.sh" # Every dependency should have env_vars.sh + run.sh executable files

    if [ -z ${CONTAINER_WAS_RUNNING} ]; then
        echo "Container ${GETH_DOCKER_CONTAINER_NAME} is not currently running"
        echo "Executing [${path}/run.sh]"
        ${path}/run.sh
        if [ $? -ne 0 ]; then
          exit 1
        fi
    else
      echo "Container ${GETH_DOCKER_CONTAINER_NAME} is already running. Not starting again."
    fi
done
############################################################

################# Prepare for tests ########################
cd $CENT_ETHEREUM_CONTRACTS_DIR
npm install

# Unlock User to Run Migration and Run it
docker run -it --net=host ethereum/client-go:$GETH_DOCKER_VERSION attach "${CENT_ETHEREUM_GETH_SOCKET}" --exec "personal.unlockAccount('0x${CENT_ETHEREUM_ACCOUNTS_MIGRATE_ADDRESS}', '${CENT_ETHEREUM_ACCOUNTS_MIGRATE_PASSWORD}')"
truffle migrate --network localgeth -f 2
status=$?

cd ${PARENT_DIR}

############################################################

################# Run Tests ################################
if [ $status -eq 0 ]; then
  statusAux=0
  for path in ${local_dir}/tests/*; do
    [ -x "${path}" ] || continue # if not an executable, skip
    ./$path
    statusAux="$(( $statusAux | $? ))"
  done
  # Store status of tests
  status=$statusAux
fi
############################################################

################# CleanUp ##################################
if [ ! -z ${CONTAINER_WAS_RUNNING} ]; then
    echo "Bringing GETH Daemon Down"
    docker rm -f geth-node

    # Cleaning extra DAG file, so we do not cache it - travis
    if [[ "X${RUN_CONTEXT}" == "Xtravis" ]];
    then
      new_dag=`ls -ltr $DATA_DIR/.ethash/* | tail -1 | awk '{print $9}' | tr -d '\n'`
      rm -Rf $new_dag
    fi
fi
############################################################

################# Propagate test status ####################
exit $status
############################################################
