#!/usr/bin/env bash

echo "Running Functional Tests"

status=$?
for d in $(go list -tags=functional ./... | grep -v vendor); do
    output="go test -v -race -coverprofile=profile.out -covermode=atomic -tags=functional $d"

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