# Centrifuge P2P node

[![Tests](https://github.com/centrifuge/go-centrifuge/actions/workflows/tests.yml/badge.svg?branch=develop)](https://github.com/centrifuge/go-centrifuge/actions/workflows/tests.yml)
[![GoDoc Reference](https://godoc.org/github.com/centrifuge/go-centrifuge?status.svg)](https://godoc.org/github.com/centrifuge/go-centrifuge)
[![codecov](https://codecov.io/gh/centrifuge/go-centrifuge/branch/develop/graph/badge.svg)](https://codecov.io/gh/centrifuge/go-centrifuge)
[![Go Report Card](https://goreportcard.com/badge/github.com/centrifuge/go-centrifuge)](https://goreportcard.com/report/github.com/centrifuge/go-centrifuge)

`go-centrifuge` is the go implementation of the Centrifuge P2P protocol. It connects to other nodes via libp2p2 and uses Centrifuge Chain for on-chain interactions.

**Getting help:** Head over to our documentation at [docs.centrifuge.io](http://docs.centrifuge.io) to learn how to setup a node and interact with it. If you have any questions, feel free to join our [discord](https://centrifuge.io/discord)

## Pre-requisites
- Go >= 1.18.x
- Node: 10.15.1
- Npm: 6.xx
- Truffle 5.1.29
- jq(https://stedolan.github.io/jq/)
- flock

## Fetch dependencies
To fetch the dependencies, run `make install-deps`.

## Install Node
To install, run `make install` will compile project to binary `centrifuge` and be placed under `GOBIN`.

Ensure `GOBIN` is under `PATH` to call the binary globally.

## Running tests
There 4 different flavours of tests in the project
- Unit tests(unit)
- Integration tests(integration)
- Environment tests(testworld): spins up multiple go-centrifuge nodes and local ethereum and centrifuge chains

To run all the tests:
- `make run-unit-tests`
- `make run-integration-tests`
- `make run-testworld-tests`

## Deploying locally
If you like to deploy all the dependencies including node, please follow below steps

### Create config file and Start centrifuge node locally:
To start centrifuge node locally, follow the steps below:
- Start the local test environment
- Run `make start-local-node`
  - This will start a local centrifuge node and create a config if not present already
  - The default config file will be placed under `~/centrifuge/testing`.
  - If you like to recreate config, then run `recreate_config=true make start-local-node`

## API definitions
Node APIs are published to swagger hub.
For the latest APIs, please see here: [APIs](https://app.swaggerhub.com/apis/centrifuge.io/cent-node/)

