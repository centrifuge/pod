#!/usr/bin/env bash

# Multiple coroutines might execute this script concurrently, the following acts as a lock.
[ "${FLOCKER}" != "$0" ] && exec env FLOCKER="$0" flock -e "$0" "$0" "$@"

CENT_CHAIN_DOCKER_START_TIMEOUT=${CENT_CHAIN_DOCKER_START_TIMEOUT:-600}
CENT_CHAIN_DOCKER_START_INTERVAL=${CENT_CHAIN_DOCKER_START_INTERVAL:-2}

CC_DOCKER_CONTAINER_WAS_RUNNING=$(docker ps -a --filter "name=cc-alice" --filter "status=running" --quiet)

echo "centchain node was running? [${CC_DOCKER_CONTAINER_WAS_RUNNING}]"
if [ -n "${CC_DOCKER_CONTAINER_WAS_RUNNING}" ]; then
    echo "Container ${CC_DOCKER_CONTAINER_NAME} is already running. Not starting again."
    exit 0;
else
    echo "Container ${CC_DOCKER_CONTAINER_NAME} is not currently running. Going to start."
fi

function wait_for_container() {
  container_name=$1
  if [ "$container_name" == "" ]; then
    echo "Please provide a docker container name."
    exit 1
  fi

  echo "Waiting for docker container '$container_name' to start up..."

  maxCount=$(( CENT_CHAIN_DOCKER_START_TIMEOUT / CENT_CHAIN_DOCKER_START_INTERVAL ))
  echo "MaxCount: $maxCount"

  count=0
  while true
  do
    validating=$(docker logs "$container_name" 2>&1 | grep 'finalized #')
    if [ "$validating" != "" ]; then
      echo "Container '$container_name' successfully started"
      break
    elif [ $count -ge $maxCount ]; then
      echo "Timeout reached while waiting for container '$container_name'"
      exit 1
    fi
    sleep "$CENT_CHAIN_DOCKER_START_INTERVAL";
    ((count++))
  done
}

cc_docker_image_tag="${PARA_DOCKER_IMAGE_TAG:-latest}"
parachain_spec="${PARA_CHAIN_SPEC:-centrifuge-local}"

export PARA_DOCKER_IMAGE_TAG=$cc_docker_image_tag
export CC_DOCKER_TAG=$cc_docker_image_tag
export PARA_CHAIN_SPEC=$parachain_spec

# Setup
PARENT_DIR=$(pwd)

cd "${PARENT_DIR}"/build/centrifuge-chain/ || exit

################## Run RelayChain #########################
"${PARENT_DIR}"/build/centrifuge-chain/scripts/init.sh start-relay-chain

wait_for_container "alice"

################## Run CentChain #########################
"${PARENT_DIR}"/build/centrifuge-chain/scripts/init.sh start-parachain-docker

wait_for_container "cc-alice"

################## Onboard ###############################

echo "sourcing in nvm"
. $NVM_DIR/nvm.sh
nvm use v17

echo "Onboarding Centrifuge Parachain ..."
DOCKER_ONBOARD=true \
./scripts/init.sh onboard-parachain

echo "Note that the Centrifuge Chain will start producing blocks when onboarding is complete"