#!/usr/bin/env bash

echo "Running Integration Tests"

status=$?
for d in $(go list ./... | grep -v vendor); do
    output="go test -v -race -coverprofile=profile.out -covermode=atomic -tags=integration $d 2>&1"

    if [ -x "$(command -v richgo)" ]; then
     output="$output | tee >(richgo testfilter)"
    fi

    if [ $? -ne 0 ]; then
      status=1
    fi

    eval "$output" | while IFS= read -r line; do printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$line"; done
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done

exit $status