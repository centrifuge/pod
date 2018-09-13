#!/usr/bin/env bash

echo "Running Unit Tests"

status=$?
for d in $(go list -tags=unit ./... | grep -v vendor); do
    output="go test -v -race -coverprofile=profile.out -covermode=atomic -tags=unit $d 2>&1"

    if [ -x "$(command -v richgo)" ]; then
     output="$output | tee >(richgo testfilter)"
    fi

    eval "$output"
    if [ $? -ne 0 ]; then
      status=1
    fi

    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

exit $status
