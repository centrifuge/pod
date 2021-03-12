#!/usr/bin/env bash

echo "Running Unit Tests"

output="go test -race -coverprofile=coverage.txt -covermode=atomic -run ${TEST_FUNCTION} -tags=unit ./... 2>&1"
if [[ -z "${TEST_FUNCTION}" ]]; then
  output="go test -race -coverprofile=coverage.txt -covermode=atomic -tags=unit ./... 2>&1"
fi

if ! eval "$output"; then
  status=1
fi

exit $status
