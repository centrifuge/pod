#!/usr/bin/env bash

echo "Running Unit Tests"

for d in $(go list ./... | grep -v vendor); do
    go test -v -coverprofile=profile.out -covermode=atomic -tags=unit $d | while IFS= read -r line; do printf '[%s] %s\n' "$(date '+%Y-%m-%d %H:%M:%S')" "$line"; done
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
