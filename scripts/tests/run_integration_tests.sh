#!/bin/bash
set -a

PARENT_DIR=`pwd`
source "${PARENT_DIR}/scripts/test-dependencies/test-ethereum/env_vars.sh"

################# Prepare for tests ########################
# Get latest Anchor Registry Address from contract json
export TEST_TIMEOUT=${TEST_TIMEOUT:-720s}
export TEST_TARGET_ENVIRONMENT=${TEST_TARGET_ENVIRONMENT:-'local'}
export CENT_CENTRIFUGENETWORK=${CENT_CENTRIFUGENETWORK:-'testing'}
export CENT_NETWORKS_TESTING_CONTRACTADDRESSES_ANCHORREGISTRY=`cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.AnchorRegistry.address' | tr -d '\n'`
export CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYFACTORY=`cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.IdentityFactory.address' | tr -d '\n'`
export CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYREGISTRY=`cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.IdentityRegistry.address' | tr -d '\n'`
echo "ANCHOR ADDRESS: ${CENT_NETWORKS_TESTING_CONTRACTADDRESSES_ANCHORREGISTRY}"
echo "IDENTITY FACTORY ADDRESS: ${CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYFACTORY}"
echo "IDENTITY REGISTRY ADDRESS: ${CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYREGISTRY}"

############################################################

echo "Running Integration Ethereum Tests against [${CENT_ETHEREUM_NODEURL}] with TIMEOUT [${TEST_TIMEOUT}]"
go test ./... -tags=ethereum -timeout ${TEST_TIMEOUT}
