#!/usr/bin/env bash
set -e
set -x

FMT_OUT="$($GOPATH/bin/goimports -d -e -l ./centrifuge)"

if [[ "$FMT_OUT" ]]; then
    echo "Go code is not formatted:"
    exit 1
fi
