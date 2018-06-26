#!/usr/bin/env bash

echo "Running Unit Tests"

for d in $(go list ./... | grep -v vendor); do
    go test -v -coverprofile=profile.out -covermode=atomic -tags=unit $d | ts '[%Y-%m-%d %H:%M:%S]'
    if [ -f profile.out ]; then
        cat profile.out >> coverage.txt
        rm profile.out
    fi
done
