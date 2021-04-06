#!/usr/bin/env bash

echo "Running Testworld"

cleanup="ls testworld/hostconfigs/* | grep testworld | grep -v README.md | tr -d : | xargs rm -rf"

eval "$cleanup"

status=$?

output="go test -timeout 30m -v -coverprofile=profile.out -covermode=atomic -tags=testworld github.com/centrifuge/go-centrifuge/testworld 2>&1"
eval "$output" | while IFS= read -r line; do printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$line"; done
if [ "${PIPESTATUS[0]}" -ne 0 ]; then
  status=1
fi

if [ -f profile.out ]; then
    cat profile.out >> coverage.txt
    rm profile.out
fi

eval "$cleanup"

exit $status
