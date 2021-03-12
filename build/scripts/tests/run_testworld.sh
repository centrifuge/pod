#!/usr/bin/env bash

echo "Running Testworld"

output="go test -timeout 30m -v -coverprofile=coverage.txt -covermode=atomic -run ${TEST_FUNCTION} -tags=testworld github.com/centrifuge/go-centrifuge/testworld 2>&1"
if [[ -z "${TEST_FUNCTION}" ]]; then
  output="go test -timeout 30m -v -coverprofile=coverage.txt -covermode=atomic -tags=testworld github.com/centrifuge/go-centrifuge/testworld 2>&1"
fi

if ! eval "$output"; then
  status=1
fi

exit $status
