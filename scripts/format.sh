#!/bin/bash
set -e
set -x

if [ -n "$(goimports -d -e -l .)" ]; then
    echo "Go code is not formatted:"
    goimports -d -e -l .
    exit 1
fi