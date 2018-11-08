#!/usr/bin/env bash
set -a

PARENT_DIR=`pwd`
source "${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/env_vars.sh"

################# Prepare for tests ########################
source "${PARENT_DIR}/build/scripts/setup_smart_contract_addresses.sh"

echo "Running Integration Tests against [${CENT_ETHEREUM_NODEURL}] with TIMEOUT [${TEST_TIMEOUT}]"

status=$?
for d in $(go list -tags=integration ./... | grep -v vendor); do
    d="github.com/centrifuge/go-centrifuge/nft"
    output="go test -race -coverprofile=profile.out -covermode=atomic -tags=integration $d 2>&1 -v"
    eval "$output"| while IFS= read -r line; do printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$line"; done
    if [ ${PIPESTATUS[0]} -ne 0 ]; then
      status=1
    fi

    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

exit $status