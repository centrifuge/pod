#!/usr/bin/env bash

CC_DOCKER_CONTAINER_NAME="cc-node"
CENT_ETHEREUM_GETH_START_TIMEOUT=${CENT_ETHEREUM_GETH_START_TIMEOUT_OVERRIDE:-600} # In Seconds, default 10 minutes
CENT_ETHEREUM_GETH_START_INTERVAL=${CENT_ETHEREUM_GETH_START_INTERVAL_OVERRIDE:-2} # In Seconds, default 2 seconds

export PARA_CHAIN_SPEC=${PARA_CHAIN_SPEC:-development-local}
# Keep image up to date accordingly
export CC_DOCKER_TAG=${CC_DOCKER_TAG:-parachain-latest}
