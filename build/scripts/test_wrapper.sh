#!/usr/bin/env bash
set -a

# Setup
local_dir="$(dirname "$0")"
PARENT_DIR=$(pwd)

# should migrate contracts
MIGRATE=${FORCE_MIGRATE:-'false'}

# should we cleanup
CLEANUP=${CLEANUP:-'false'}

# should run tests
RUN_TESTS=${RUN_TESTS:-'true'}


if [ "$RUN_TESTS" == 'true' ] ; then
  args=( "$@" )
  if [ $# == 0 ]; then
    args=(  unit cmd testworld integration )
    MIGRATE=true
  elif [ $# == 1 ] && [ "${args[0]}" == "unit" ]; then
    MIGRATE=false
  else
    MIGRATE=true
  fi
fi

GETH_DOCKER_CONTAINER_NAME="geth-node"
CC_DOCKER_CONTAINER_NAME="cc-node"
BRIDGE_CONTAINER_NAME="bridge"
GETH_DOCKER_CONTAINER_WAS_RUNNING=`docker ps -a --filter "name=${GETH_DOCKER_CONTAINER_NAME}" --filter "status=running" --quiet`
CC_DOCKER_CONTAINER_WAS_RUNNING=`docker ps -a --filter "name=${CC_DOCKER_CONTAINER_NAME}" --filter "status=running" --quiet`
BRIDGE_DOCKER_CONTAINER_WAS_RUNNING=`docker ps -a --filter "name=${BRIDGE_CONTAINER_NAME}" --filter "status=running" --quiet`

################# Run Ethereum and Centrifuge chain Nodes #########################
for path in ${local_dir}/test-dependencies/test-*; do
    [ -d "${path}" ] || continue # if not a directory, skip
    source "${path}/env_vars.sh" # Every dependency should have env_vars.sh + run.sh executable files
    ${path}/run.sh
    if [ $? -ne 0 ]; then
        exit 1
    fi
done

################# Migrate contracts ########################
if [ "$MIGRATE" == 'true' ]; then
  rm -f /tmp/migration.log
  docker rm -f bridge
  BRIDGE_DOCKER_CONTAINER_WAS_RUNNING=
  echo "Running migrations? [${MIGRATE}]"
  echo "Logging to /tmp/migration.log..."
  "${PARENT_DIR}"/build/scripts/migrate.sh &> /tmp/migration.log
  if [ $? -ne 0 ]; then
    echo "migrations failed"
    cat /tmp/migration.log
    exit 1
  fi
  rm -f /tmp/migration.log
  ## adding this env here as well since the envs from previous step(child script) is not imported
  export MIGRATION_RAN=true

  ################# deploy bridge########################
  ## delete any stale bridge containers
  path=${local_dir}/test-dependencies/bridge
  source "${path}/env_vars.sh"
  ${path}/run.sh
  if [ $? -ne 0 ]; then
      exit 1
  fi
fi

################# Run Tests ################################
if [ "$RUN_TESTS" == 'true' ] ; then
  # Code coverage is stored in coverage.txt
  echo "" > coverage.txt

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

  ################# Propagate test status ####################
  echo "The test suite overall is exiting with status [$status]"
  exit $status
  ############################################################
fi

################# CleanUp ##################################
if [ "${CLEANUP}" == "true" ]; then
  echo "Bringing GETH Daemon Down"
  docker rm -f geth-node
  echo "Bringing Centrifuge Chain down"
  docker rm -f cc-node
  echo "Bringing bridge down..."
  docker rm -f bridge
fi
