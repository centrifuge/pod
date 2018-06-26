#!/usr/bin/env bash

set -a

PARENT_DIR=`pwd`
source "${PARENT_DIR}/scripts/test-dependencies/test-ethereum/env_vars.sh"

################# Prepare for tests ########################
# Get latest Anchor Registry Address from contract json
export TEST_TIMEOUT=${TEST_TIMEOUT:-600s}
export TEST_TARGET_ENVIRONMENT=${TEST_TARGET_ENVIRONMENT:-'local'}
export CENT_CENTRIFUGENETWORK=${CENT_CENTRIFUGENETWORK:-'testing'}

## Making Env Var Name dynamic
cent_upper_network=`echo $CENT_CENTRIFUGENETWORK | awk '{print toupper($0)}'`
temp1="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_ANCHORREGISTRY"
printf -v $temp1 `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.AnchorRegistry.address' | tr -d '\n'`
temp2="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_IDENTITYFACTORY"
printf -v $temp2 `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.IdentityFactory.address' | tr -d '\n'`
temp3="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_IDENTITYREGISTRY"
printf -v $temp3 `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.IdentityRegistry.address' | tr -d '\n'`
export $temp1
export $temp2
export $temp3
vtemp1=$(eval "echo \"\$$temp1\"")
vtemp2=$(eval "echo \"\$$temp2\"")
vtemp3=$(eval "echo \"\$$temp3\"")
#

echo "ANCHOR REGISTRY ADDRESS: ${vtemp1}"
echo "IDENTITY REGISTRY ADDRESS: ${vtemp3}"
echo "IDENTITY FACTORY ADDRESS: ${vtemp2}"
#
#############################################################
#
echo "Running Integration Ethereum Tests against [${CENT_ETHEREUM_NODEURL}] with TIMEOUT [${TEST_TIMEOUT}]"
for d in $(go list ./... | grep -v vendor); do
    go test -v -coverprofile=profile.out -covermode=atomic -tags=ethereum -timeout ${TEST_TIMEOUT} $d |  while IFS= read -r line; do printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$line"; done
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
