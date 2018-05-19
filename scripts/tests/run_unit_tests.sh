#!/usr/bin/env bash

echo "Running Unit Tests"
go test ./... -tags=unit -race -coverprofile=coverage.txt -covermode=atomic
