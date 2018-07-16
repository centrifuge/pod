#!/usr/bin/env bash

set -e

#docker build -t centrifugeio/go-centrifuge:ubuntu -f Dockerfile.ubuntu .

mkdir -p build/linux
cid=$(docker create centrifugeio/go-centrifuge:20180716-9265ad0)
docker cp $cid:/root/centrifuge build/linux/
docker rm -v $cid

#sha_commit=`git rev-parse --short HEAD`
#
#tar -zcvf cent-api-${sha_commit}.tar.gz -C build/linux .