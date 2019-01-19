#!/usr/bin/env bash

set -a

################# Prepare for tests ########################
echo "Running CMD Tests"

status=$?
for d in $(go list -tags=cmd ./... | grep cmd | grep -v vendor); do
    output="go test -coverprofile=profile.out -covermode=atomic -tags=cmd $d 2>&1"
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