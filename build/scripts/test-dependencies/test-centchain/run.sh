#!/usr/bin/env bash

echo "centchain node was running? [${CC_DOCKER_CONTAINER_WAS_RUNNING}]"
if [ -n "${CC_DOCKER_CONTAINER_WAS_RUNNING}" ]; then
    echo "Container ${CC_DOCKER_CONTAINER_NAME} is already running. Not starting again."
    exit 0;
else
    echo "Container ${CC_DOCKER_CONTAINER_NAME} is not currently running. Going to start."
fi

# Setup
PARENT_DIR=$(pwd)
local_dir="$(dirname "$0")"
source "${local_dir}/env_vars.sh"

################## Run RelayChain #########################
cd "${PARENT_DIR}"/build/centrifuge-chain || exit
## Tweaking network
default_network=$(cat docker-compose-local-relay.yml | grep "name: docker_default")
if [[ $default_network == "" ]]; then
cat <<EOT >> docker-compose-local-relay.yml
networks:
  default:
    external:
      name: docker_default
EOT
fi

./scripts/init.sh start-relay-chain

echo "Waiting for Relay Chain to Start Up ..."
maxCount=$(( CENT_ETHEREUM_GETH_START_TIMEOUT / CENT_ETHEREUM_GETH_START_INTERVAL ))
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
  sleep "$CENT_ETHEREUM_GETH_START_INTERVAL";
  ((count++))
done

################## Run CentChain #########################
## Centrifuge Chain local Development testnet
## Tweaking network
default_network=$(cat docker-compose-local-chain.yml | grep "name: docker_default")
if [[ $default_network == "" ]]; then
cat <<EOT >> docker-compose-local-chain.yml
networks:
  default:
    external:
      name: docker_default
EOT
fi

./scripts/init.sh start-parachain-docker

echo "Waiting for Centrifuge Chain to Start Up ..."
maxCount=$(( CENT_ETHEREUM_GETH_START_TIMEOUT / CENT_ETHEREUM_GETH_START_INTERVAL ))
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
  sleep "$CENT_ETHEREUM_GETH_START_INTERVAL";
  ((count++))
done

nvm use v17
echo "Onboarding Centrifuge Parachain ..."
DOCKER_ONBOARD=true PARA_CHAIN_SPEC=development-local ./scripts/init.sh onboard-parachain

echo "Not waiting for Centrifuge Chain to start producing blocks since geth needs to start and migrate needs to happen"

cd "${PARENT_DIR}" || exit