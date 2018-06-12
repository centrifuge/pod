#!/usr/bin/env bash

export CENT_ETHEREUM_INTERVALRETRY='100ms'

echo "Running Unit Tests"
go test ./... -tags=unit
