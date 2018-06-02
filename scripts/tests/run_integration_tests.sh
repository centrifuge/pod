#!/bin/bash
set -a

PARENT_DIR=`pwd`
source "${PARENT_DIR}/scripts/test-dependencies/test-ethereum/env_vars.sh"

################# Prepare for tests ########################
# Get latest Anchor Registry Address from contract json
export CENT_ANCHOR_ETHEREUM_ANCHORREGISTRYADDRESS=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/AnchorRegistry.json | jq -r --arg NETWORK_ID "${NETWORK_ID}" '.networks[$NETWORK_ID].address' | tr -d '\n'`
export CENT_IDENTITY_ETHEREUM_IDENTITYFACTORYADDRESS=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/IdentityFactory.json | jq -r --arg NETWORK_ID "${NETWORK_ID}" '.networks[$NETWORK_ID].address' | tr -d '\n'`
export CENT_IDENTITY_ETHEREUM_IDENTITYREGISTRYADDRESS=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/IdentityRegistry.json | jq -r --arg NETWORK_ID "${NETWORK_ID}" '.networks[$NETWORK_ID].address' | tr -d '\n'`
export CENT_IDENTITY_ETHEREUM_IDENTITYADDRESS=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/Identity.json | jq -r --arg NETWORK_ID "${NETWORK_ID}" '.networks[$NETWORK_ID].address' | tr -d '\n'`
echo "ANCHOR ADDRESS: ${CENT_ANCHOR_ETHEREUM_ANCHORREGISTRYADDRESS}"
echo "IDENTITY FACTORY ADDRESS: ${CENT_IDENTITY_ETHEREUM_IDENTITYFACTORYADDRESS}"
echo "IDENTITY REGISTRY ADDRESS: ${CENT_IDENTITY_ETHEREUM_IDENTITYREGISTRYADDRESS}"
############################################################

echo "Running Integration Ethereum Tests against [${CENT_ETHEREUM_GETH_SOCKET}]"
#go test -v ./... -tags=ethereum -timeout 60s
