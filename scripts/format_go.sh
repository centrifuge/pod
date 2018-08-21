#!/bin/bash
set -e
set -x

if [ -n "$($GOPATH/bin/goimports -d -e -l .)" ]; then
    echo "Go code is not formatted:"
    $GOPATH/bin/goimports -d -e -l ./centrifuge
    exit 1
fi