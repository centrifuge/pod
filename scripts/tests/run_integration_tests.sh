#!/bin/bash
set -a

PARENT_DIR=`pwd`
source "${PARENT_DIR}/scripts/test-dependencies/test-ethereum/env_vars.sh"

################# Prepare for tests ########################
# Get latest Anchor Registry Address from contract json
export CENT_NETWORKS_TESTING_CONTRACTADDRESSES_ANCHORREGISTRY=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/AnchorRegistry.json | jq -r --arg NETWORK_ID "${NETWORK_ID}" '.networks[$NETWORK_ID].address' | tr -d '\n'`
export CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYFACTORY=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/IdentityFactory.json | jq -r --arg NETWORK_ID "${NETWORK_ID}" '.networks[$NETWORK_ID].address' | tr -d '\n'`
export CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYREGISTRY=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/IdentityRegistry.json | jq -r --arg NETWORK_ID "${NETWORK_ID}" '.networks[$NETWORK_ID].address' | tr -d '\n'`
echo "ANCHOR ADDRESS: ${CENT_NETWORKS_TESTING_CONTRACTADDRESSES_ANCHORREGISTRY}"
echo "IDENTITY FACTORY ADDRESS: ${CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYFACTORY}"
echo "IDENTITY REGISTRY ADDRESS: ${CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYREGISTRY}"

############################################################

echo "Running Integration Ethereum Tests against [${CENT_ETHEREUM_NODEURL}]"
go test ./... -tags=ethereum -timeout 90s
