# Centrifuge OS node 

[![Build Status](https://travis-ci.com/centrifuge/go-centrifuge.svg?token=Sbf68xBZUZLMB3kGTKcX&branch=master)](https://travis-ci.com/centrifuge/go-centrifuge)
[![GoDoc Reference](https://godoc.org/github.com/centrifuge/go-centrifuge?status.svg)](https://godoc.org/github.com/centrifuge/go-centrifuge)
[![codecov](https://codecov.io/gh/centrifuge/go-centrifuge/branch/develop/graph/badge.svg)](https://codecov.io/gh/centrifuge/go-centrifuge)
[![Go Report Card](https://goreportcard.com/badge/github.com/centrifuge/go-centrifuge)](https://goreportcard.com/report/github.com/centrifuge/go-centrifuge)

`go-centrifuge` is the go implementation of the Centrifuge OS interacting with the peer to peer network and our Ethereum smart contracts. 

**Getting help:** Head over to our developer documentation at [developer.centrifuge.io](http://developer.centrifuge.io) to learn how to setup a node and interact with it. If you have any questions, feel free to join our [slack channel](https://join.slack.com/t/centrifuge-io/shared_invite/enQtNDYwMzQ5ODA3ODc0LTU4ZjU0NDNkOTNhMmUwNjI2NmQ2MjRiNzA4MGIwYWViNTkxYzljODU2OTk4NzM4MjhlOTNjMDAwNWZkNzY2YWY) 

**DISCLAIMER:** The code released here presents a very early alpha version that should not be used in production and has not been audited. Use this at your own risk.

## Table of Contents
- [Installing pre-requisites](#installing-pre-requisites)
 - [Linux](#linux)
    - [Install Docker Compose](#install-docker-compose)
 - [Mac](#mac)
    - [Install Docker Compose](#install-docker-compose-1)
- [Build](#build)
- [Install](#install)
- [Running Tests](#running-tests)
- [Running tests continuously while developing](#running-tests-continuously-while-developing)
- [Run a Geth node locally or Rinkeby environments](#run-a-geth-node-locally-or-rinkeby-environments)
    - [Run as local node in dev mode](#run-as-local-node-with-mining-enabled)
    - [Run local peer connected to Rinkeby](#run-local-peer-connected-to-rinkeby)
    - [Checking on your local geth node](#checking-on-your-local-geth-node)
    - [Attaching to your local geth node](#attaching-to-your-local-geth-node)
- [Run Centrifuge Chain locally in dev mode](#run-centrifuge-chain-locally-in-dev-mode)
- [Run Integration Tests against Local/Rinkeby Environments](#run-integration-tests-against-localintegrationrinkeby-environments)
 - [Configure local dev node run integration/functional tests](#configure-local-mining--run-integrationfunctional-tests)
 - [Configure node to point to integration run integration/functional tests](#configure-node-to-point-to-integration--run-integrationfunctional-tests)
 - [Configure node to point to rinkeby run integration/functional tests](#configure-node-to-point-to-rinkeby--run-integrationfunctional-tests)
 - [Configure node to point to infura-rinkeby run integration/functional tests](#configure-node-to-point-to-infura-rinkeby--run-integrationfunctional-tests)
- [Ethereum Contract Bindings](#ethereum-contract-bindings)
- [Protobufs bindings](#protobufs-bindings)


## Installing pre-requisites
### Linux
```bash
# Install Go

sudo apt-get update
sudo apt-get -y upgrade

#download go using this command
wget https://dl.google.com/go/go1.11.4.linux-amd64.tar.gz

#extract the archive and move it to /usr/local folder
sudo tar -xvf go1.11.4.linux-amd64.tar.gz
sudo mv go /usr/local

#cd to ~/.profile and add the following lines to the end.
export GOROOT=/usr/local/go
export GOPATH=$HOME/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH

# execute this command to use go in the current shell. 
source ~/.profile

# install jq
sudo apt-get install jq

# install truffle framework
npm install -g truffle@4.0.4

# install Dep
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
```

#### Install Docker Compose
```bash
# Run this command to download the latest version of Docker Compose
sudo curl -L https://github.com/docker/compose/releases/download/1.22.0/docker-compose-$(uname -s)-$(uname -m) -o /usr/local/bin/docker-compose

# Apply executable permissions to the binary
sudo chmod +x /usr/local/bin/docker-compose
```

### Mac
```bash
# install jq
brew install jq

# install truffle framework
npm install -g truffle@4.0.4

# install Dep
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh
```

#### Install Docker Compose
Make sure you have docker-compose installed, usually comes bundled with Mac OS Docker. Otherwise: https://docs.docker.com/compose/install/


## Build

Build & install the Centrifuge OS Node
```bash
mkdir -p $GOPATH/src/github.com/centrifuge/go-centrifuge/
cd $GOPATH/src/github.com/centrifuge/
git clone git@github.com:centrifuge/go-centrifuge.git $GOPATH/src/github.com/centrifuge/go-centrifuge
cd go-centrifuge
make install
```
Check whether everything is fine by running tests

```bash
go test --tags="unit" ./...
go test --tags="integration" ./...
```

## Running Tests

Install packages and dependencies
```bash
dep ensure
```
Run only unit tests
```bash
go test --tags="unit" ./...
```

Run only integration tests:
```bash
go test --tags="integration" ./...
```

For Testworld tests, please refer to the [Testworld README](testworld/README.md)

To run integration/functional tests a few other components need to be set up.
- Geth node needs to be up and running
- Contracts need to be deployed
    - Run contract migration (fetched by ENV_VAR CENT_ETHEREUM_CONTRACTS_DIR under `build/scripts/test-dependencies/test-ethereum/env_vars.sh` )
- Local account keys need to be set and able to call the right contracts on Ethereum

To do this setup + run all the tests (unit, integration, functional) use the `test_wrapper.sh` script.

Run the whole test-suite with
```bash
./build/scripts/test_wrapper.sh
```

Please note that this testing script is intended for CI testing. Please use the `go test` commands for unit and integration testing.

### Running tests continuously while developing

If you want to run tests continuously when a file changes, you can use [reflex](https://github.com/cespare/reflex):
```bash
go get github.com/cespare/reflex
```

Then run (only for unit tests). It is a good idea to exclude the `vendor` and ``.idea` folders from scanning by reflex as it hogs a lot of resources for no good reason.
```bash
reflex -R '(^|/)vendor/|(^|/)\\.idea/' -- go test ./centrifuge/... -tags=unit
```

Or run for specific tests only:
```bash
reflex -R '(^|/)vendor/|(^|/)\\.idea/' -- go test ./centrifuge/invoice/... -tags=unit
```

## Run Centrifuge Chain locally in dev mode
For development, we use Docker Compose locally to run the Centrifuge Chain. It comes with a set of preconfigured accounts to be used. 

`./build/scripts/docker/run.sh ccdev`

For more info: https://github.com/centrifuge/centrifuge-chain

## Run a Geth node locally or Rinkeby environments

For development, we make use of Docker Compose locally as it is easy and clear to bundle volume and environment configurations:
Docker Compose files live in `./build/scripts/docker`

#### Run as local node in dev mode
Then run the local node via `./build/scripts/docker/run.sh dev`
By default it uses:
- ETH_DATADIR=${HOME}/Library/Ethereum
- RPC_PORT=9545
- WS_PORT=9546


#### Run local peer connected to Rinkeby
Let's run the rinkeby local node:
`./build/scripts/docker/run.sh rinkeby`
By default it uses:
- ETH_DATADIR=${HOME}/Library/Ethereum
- RPC_PORT=9545

Override those when needed

Let it catch up for a while until is fully synced with the remote peer

## Run Integration Tests against Local/Rinkeby Environments

### Configure local dev node + run integration/functional tests
  - Remove running container if any:
    - docker rm -f geth-node
  - Clear up ~/Library/Ethereum/8383 folder (keep in mind this will clear up all previous data you had before)
    - rm -Rf ~/Library/Ethereum/8383
  - In go-centrifuge project run:
    - ./build/scripts/docker/run.sh dev
  - Run contract migration to generate local contract address artifact:
    - In centrifuge-ethereum-contracts project:
      - ./build/scripts/migrate.sh localgeth
      - Verify that ./deployments/local.json has been generated (note that local.json is excluded in .gitignore)
  - Run tests:
    - To run only integration tests:
      - ./build/scripts/tests/run_integration_tests.sh

### Configure node to point to rinkeby + run integration/functional tests
  - Remove running container if any:
    - docker rm -f geth-node
  - In go-centrifuge project run:
    - ./build/scripts/docker/run.sh rinkeby
    - Wait until node is in sync with remote peer (1-2 hours):
      - geth attach http://localhost:9545 --exec "net.peerCount" > 0 (rinkeby takes additional time to sync as it needs a peer to pull from, and has shortage of full node peers)
      - geth attach http://localhost:9545 --exec "eth.syncing" -> false
  - Run tests:
    - To run only integration tests:
      - CENT_CENTRIFUGENETWORK='russianhill' TEST_TARGET_ENVIRONMENT='rinkeby' CENT_ETHEREUM_ACCOUNTS_MAIN_KEY='$JSON_KEY' CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD="$PASS" CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS="$ADDR" ./build/scripts/tests/run_integration_tests.sh

### Configure node to point to infura-rinkeby + run integration/functional tests ####
  - Run tests:
    - To run only integration tests:
      - CENT_ETHEREUM_TXPOOLACCESSENABLED=false CENT_ETHEREUM_NODEURL='wss://rinkeby.infura.io/ws/MtCWERMbJtnqPKI8co84' CENT_CENTRIFUGENETWORK='russianhill' TEST_TARGET_ENVIRONMENT='rinkeby' CENT_ETHEREUM_ACCOUNTS_MAIN_KEY='$JSON_KEY' CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD="$PASS" CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS="$ADDR" ./build/scripts/tests/run_integration_tests.sh

## Ethereum Contract Bindings

To create the go bindings for the deployed truffle contract, use the following command:

```bash
abigen --abi abi/AnchorRepository.abi --pkg anchor --type EthereumAnchorRepositoryContract --out ${GOPATH}/src/github.com/centrifuge/go-centrifuge/anchor/ethereum_anchor_repository_contract.go
```

and then copy the `ethereum_anchor_registry_contract.go` file to `anchors/`. You will also need to modify the file to add the following imports:

```go,
import(
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)
```

## Protobufs bindings

Create any new `.proto` files in its own package under `protobufs` folder.
Generating go bindings and swagger with the following command

```bash
make proto-all
```
