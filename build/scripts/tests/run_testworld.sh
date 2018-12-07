#!/usr/bin/env bash

echo "Running Testworld"

status=$?

output="go test -race -coverprofile=profile.out -covermode=atomic -tags=testworld ./testworld 2>&1"
eval "$output" | while IFS= read -r line; do printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$line"; done
if [ ${PIPESTATUS[0]} -ne 0 ]; then
  status=1
fi

if [ -f profile.out ]; then
    cat profile.out >> coverage.txt
    rm profile.out
fi


exit $status
