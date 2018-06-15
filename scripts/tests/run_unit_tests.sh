#!/usr/bin/env bash

echo "Running Unit Tests"
go test ./... -tags=unit
