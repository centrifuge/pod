#!/usr/bin/env bash

set -a

PARENT_DIR=`pwd`
source "${PARENT_DIR}/scripts/test-dependencies/test-ethereum/env_vars.sh"

################# Prepare for tests ########################
source "${PARENT_DIR}/scripts/setup_smart_contract_addresses.sh"


############################################################

echo "Running Functional Ethereum Tests against [${CENT_ETHEREUM_NODEURL}] with TIMEOUT [${TEST_TIMEOUT}]"

status=$?
for d in $(go list -tags=ethereum ./... | grep -v vendor); do
    output=$(go test -v -race -coverprofile=profile.out -covermode=atomic -tags=ethereum -timeout ${TEST_TIMEOUT} $d)
    if [ $? -ne 0 ]; then
      status=1
    fi
     echo "${output}" | while IFS= read -r line; do printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$line"; done
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

exit $status