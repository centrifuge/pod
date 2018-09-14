#!/usr/bin/env bash
set -e
set -x

mkdir -p ~/bin
PROTOTOOL_VERSION=$PROTOTOOL_VERSION ./scripts/install_prototool.sh
status=$?

go get golang.org/x/tools/cmd/goimports
go get -u github.com/kyoh86/richgo
go get github.com/grpc-ecosystem/grpc-gateway/
go get github.com/CentrifugeInc/centrifuge-protobufs/
exit $status
