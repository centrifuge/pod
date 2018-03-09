#!/bin/bash

PARENT_DIR=`pwd`
source "${PARENT_DIR}/scripts/test-dependencies/test-ethereum/env_vars.sh"

################# Prepare for tests ########################
# Get latest Anchor Registry Address from contract json
export CENT_ANCHOR_ETHEREUM_ANCHORREGISTRYADDRESS=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/AnchorRegistry.json | jq -r --arg NETWORK_ID "${NETWORK_ID}" '.networks[$NETWORK_ID].address' | tr -d '\n'`
echo "ANCHOR ADDRESS: ${CENT_ANCHOR_ETHEREUM_ANCHORREGISTRYADDRESS}"
############################################################

echo "Running Integration Ethereum Tests against IPC [${CENT_ETHEREUM_GETHIPC}]"
go test ./... -tags=ethereum