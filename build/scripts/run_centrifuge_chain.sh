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

# Setup
PARENT_DIR=$(pwd)

mkdir -p /tmp/centrifuge-pod/res
mkdir -p /tmp/centrifuge-pod/deps
cp "${PARENT_DIR}"/build/centrifuge-chain/docker/docker-compose-local-relay.yml /tmp/centrifuge-pod/deps/
cp "${PARENT_DIR}"/build/centrifuge-chain/docker/docker-compose-local-chain.yml /tmp/centrifuge-pod/deps/
cp "${PARENT_DIR}"/build/centrifuge-chain/res/rococo-local.json /tmp/centrifuge-pod/res/
docker network inspect docker_default
if [ $? -ne 0 ]; then
  docker network create docker_default
fi

################## Run RelayChain #########################
cd "${PARENT_DIR}"/build/centrifuge-chain || exit
## Tweaking network
default_network=$(cat /tmp/centrifuge-pod/deps/docker-compose-local-relay.yml | grep "name: docker_default")
if [[ $default_network == "" ]]; then
cat <<EOT >> /tmp/centrifuge-pod/deps/docker-compose-local-relay.yml
networks:
  default:
    external:
      name: docker_default
EOT
fi

docker-compose -f /tmp/centrifuge-pod/deps/docker-compose-local-relay.yml up -d

echo "Waiting for Relay Chain to Start Up ..."
maxCount=$(( CENT_CHAIN_DOCKER_START_TIMEOUT / CENT_CHAIN_DOCKER_START_INTERVAL ))
echo "MaxCount: $maxCount"
count=0
while true
do
  validating=$(docker logs alice 2>&1 | grep 'finalized #')
  if [ "$validating" != "" ]; then
    echo "RelayChain successfully started"
    break
  elif [ $count -ge $maxCount ]; then
    echo "Timeout Starting out RelayChain"
    exit 1
  fi
  sleep "$CENT_CHAIN_DOCKER_START_INTERVAL";
  ((count++))
done

################## Run CentChain #########################
## Centrifuge Chain local Development testnet
## Tweaking network
default_network=$(cat /tmp/centrifuge-pod/deps/docker-compose-local-chain.yml | grep "name: docker_default")
if [[ $default_network == "" ]]; then
cat <<EOT >> /tmp/centrifuge-pod/deps/docker-compose-local-chain.yml
networks:
  default:
    external:
      name: docker_default
EOT
fi

# Temporary fix until https://github.com/centrifuge/centrifuge-chain/pull/1644 is done.
CC_DOCKER_TAG=main-ccbc3a1-23-12-07 \
PARA_CHAIN_SPEC=development-local \
docker-compose -f /tmp/centrifuge-pod/deps/docker-compose-local-chain.yml up -d cc_alice

echo "Waiting for Centrifuge Chain to Start Up ..."
maxCount=$(( CENT_CHAIN_DOCKER_START_TIMEOUT / CENT_CHAIN_DOCKER_START_INTERVAL ))
echo "MaxCount: $maxCount"
count=0
while true
do
  validating=$(docker logs cc-alice 2>&1 | grep 'finalized #')
  if [ "$validating" != "" ]; then
    echo "CentChain successfully started"
    break
  elif [ $count -ge $maxCount ]; then
    echo "Timeout Starting out CentChain"
    exit 1
  fi
  sleep "$CENT_CHAIN_DOCKER_START_INTERVAL";
  ((count++))
done

echo "sourcing in nvm"
. $NVM_DIR/nvm.sh
nvm use v17

echo "Onboarding Centrifuge Parachain ..."
DOCKER_ONBOARD=true \
PARA_CHAIN_SPEC=development-local \
./scripts/init.sh onboard-parachain

echo "Note that the Centrifuge Chain will start producing blocks when onboarding is complete"