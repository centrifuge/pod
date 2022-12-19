# Centrifuge POD (Private Off-chain Data)

[![Tests](https://github.com/centrifuge/go-centrifuge/actions/workflows/tests.yml/badge.svg?branch=develop)](https://github.com/centrifuge/go-centrifuge/actions/workflows/tests.yml)
[![GoDoc Reference](https://godoc.org/github.com/centrifuge/go-centrifuge?status.svg)](https://godoc.org/github.com/centrifuge/go-centrifuge)
[![codecov](https://codecov.io/gh/centrifuge/go-centrifuge/branch/develop/graph/badge.svg)](https://codecov.io/gh/centrifuge/go-centrifuge)
[![Go Report Card](https://goreportcard.com/badge/github.com/centrifuge/go-centrifuge)](https://goreportcard.com/report/github.com/centrifuge/go-centrifuge)

`go-centrifuge` is the go implementation of the Centrifuge P2P protocol. It connects to other nodes via `libp2p` and uses the [Centrifuge Chain](https://github.com/centrifuge/centrifuge-chain) for on-chain interactions.

**Getting help:** Head over to our documentation at [docs.centrifuge.io](http://docs.centrifuge.io) to learn how to setup the POD and interact with it. If you have any questions, feel free to join our [discord](https://centrifuge.io/discord)

## API definitions
Node APIs are published to swagger hub.
For the latest APIs, please see here: [APIs](https://app.swaggerhub.com/apis/centrifuge.io/cent-node/)

# Table of Contents
1. [Build](#build)
2. [Test](#test)
3. [Run](#run)
4. [Description](docs/README.md)

## Build

### Pre-requisites
- Go >= 1.18.x
- jq(https://stedolan.github.io/jq/)
- flock

### Fetch dependencies
To fetch the dependencies, run `make install-deps`.

### Install POD
To install, run `make install` will compile project to binary `centrifuge` and be placed under `GOBIN`.

**NOTE** - Ensure `GOBIN` is under `PATH` to call the binary globally.

## Test
There 3 different flavours of tests in the project
- Unit tests(unit)
- Integration tests(integration)
- Environment tests(testworld)

To run all the tests:
- `make run-unit-tests`
- `make run-integration-tests`
- `make run-testworld-tests`

## Run
If you like to run all the dependencies including the POD, please follow below steps
- Run `make start-local-env` - this will start the [Centrifuge Chain](https://github.com/centrifuge/centrifuge-chain) and its required services. 
- Run `make start-local-pod`
  - This will start a local centrifuge POD and create a config if not present already
  - The default config file will be placed under `/tmp/go-centrifuge/testing`.
  - If you like to recreate config, then run `recreate_config=true make start-local-node`

