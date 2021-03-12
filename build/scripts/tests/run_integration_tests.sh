#!/usr/bin/env bash

set -a

PARENT_DIR=`pwd`
source "${PARENT_DIR}/build/scripts/test-dependencies/test-ethereum/env_vars.sh"

################# Prepare for tests ########################
echo "Running Integration Tests against [${CENT_ETHEREUM_NODEURL}] with TIMEOUT [${TEST_TIMEOUT}]"

output="go test -race -coverprofile=coverage.txt -covermode=atomic -run ${TEST_FUNCTION} -tags=integration ./... 2>&1"
if [[ -z "${TEST_FUNCTION}" ]]; then
  output="go test -race -coverprofile=coverage.txt -covermode=atomic -tags=integration ./... 2>&1"
fi

if ! eval "$output"; then
  status=1
fi

exit $status
