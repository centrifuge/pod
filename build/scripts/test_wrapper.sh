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
CC_DOCKER_CONTAINER_NAME="cc-node"
GETH_DOCKER_CONTAINER_WAS_RUNNING=`docker ps -a --filter "name=${GETH_DOCKER_CONTAINER_NAME}" --filter "status=running" --quiet`
CC_DOCKER_CONTAINER_WAS_RUNNING=`docker ps -a --filter "name=${CC_DOCKER_CONTAINER_NAME}" --filter "status=running" --quiet`

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
${PARENT_DIR}/build/scripts/migrate.sh
status=$?

############################################################

################# Run Tests ################################
args=( "$@" )
if [[ $# == 0 ]]; then
        args=(  unit cmd testworld integration )
fi

if [[ ${status} -eq 0 ]]; then
  statusAux=0
  for path in ${local_dir}/tests/*; do
    [[ -x "${path}" ]] || continue # if not an executable, skip

    for arg in "${args[@]}"; do
        if [[ ${path} == *$arg* ]]; then
            echo "Executing test suite [${path}]"
            ./$path
            statusAux="$(( $statusAux | $? ))"
            continue
        fi
    done
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
