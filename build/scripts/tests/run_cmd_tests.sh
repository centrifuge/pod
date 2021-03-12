#!/usr/bin/env bash

set -a

################# Prepare for tests ########################
echo "Running CMD Tests"

output="go test -race -coverprofile=coverage.txt -covermode=atomic -run ${TEST_FUNCTION} -tags=cmd ./... 2>&1"
if [[ -z "${TEST_FUNCTION}" ]]; then
  output="go test -race -coverprofile=coverage.txt -covermode=atomic -tags=cmd ./... 2>&1"
fi

if ! eval "$output"; then
  status=1
fi

exit $status
