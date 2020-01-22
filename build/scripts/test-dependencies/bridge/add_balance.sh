#!/usr/bin/env bash

set -e

export MIGRATE_ADDRESS=${MIGRATE_ADDRESS:-'0x89b0a86583c4444acfd71b463e0d3c55ae1412a5'}
export MIGRATE_PASSWORD=${MIGRATE_PASSWORD:-''}
export DEV_BRIDGE_ADDRESS=${DEV_BRIDGE_ADDRESS:-'0x11E296f2F658ab6892E8F4dceF696C0249486B8b'}

echo "Adding some balance to Bridge Address [${DEV_BRIDGE_ADDRESS}]"
docker run --net=host ethereum/client-go:latest attach http://localhost:9545 --exec "personal.unlockAccount('${MIGRATE_ADDRESS}', '${MIGRATE_PASSWORD}', 500)"
docker run --net=host ethereum/client-go:latest attach http://localhost:9545 --exec "eth.sendTransaction({from:'${MIGRATE_ADDRESS}', to:'${DEV_BRIDGE_ADDRESS}', value: web3.toWei(10, 'ether'), gas:21000});"
