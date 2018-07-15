#!/usr/bin/env bash

echo "Running Unit Tests"

status=$?
for d in $(go list ./... | grep -v vendor); do
    output=$(go test -coverprofile=profile.out -covermode=atomic -tags=unit $d 2>&1)
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
