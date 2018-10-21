# Centrifuge OS node 

[![Build Status](https://travis-ci.com/centrifuge/go-centrifuge.svg?token=Sbf68xBZUZLMB3kGTKcX&branch=master)](https://travis-ci.com/centrifuge/go-centrifuge)

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
 - [Troubleshooting functional test setup](#troubleshooting-functional-test-setup)
 - [Running tests continuously while developing](#running-tests-continuously-while-developing)
- [Run a Geth node locally or Rinkeby environments](#run-a-geth-node-locally-or-rinkeby-environments)
    - [Run as local node with mining enabled](#run-as-local-node-with-mining-enabled)
    - [Run local peer connected to Rinkeby](#run-local-peer-connected-to-rinkeby)
    - [Checking on your local geth node](#checking-on-your-local-geth-node)
    - [Attaching to your local geth node](#attaching-to-your-local-geth-node)
- [Run Integration Tests against Local/Integration/Rinkeby Environments](#run-integration-tests-against-localintegrationrinkeby-environments)
 - [Configure local mining run integration/functional tests](#configure-local-mining--run-integrationfunctional-tests)
 - [Configure node to point to integration run integration/functional tests](#configure-node-to-point-to-integration--run-integrationfunctional-tests)
 - [Configure node to point to rinkeby run integration/functional tests](#configure-node-to-point-to-rinkeby--run-integrationfunctional-tests)
 - [Configure node to point to infura-rinkeby run integration/functional tests](#configure-node-to-point-to-infura-rinkeby--run-integrationfunctional-tests)
- [Ethereum Contract Bindings](#ethereum-contract-bindings)
- [Protobufs bindings](#protobufs-bindings)


## Installing pre-requisites
### Linux
```bash
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
cd $GOPATH/src/github.com/centrifuge/go-centrifuge
make install
```

## Install
```bash
mkdir -p $GOPATH/src/github.com/centrifuge/go-centrifuge/
git clone git@github.com:centrifuge/go-centrifuge.git $GOPATH/src/github.com/centrifuge/go-centrifuge

# initialize your local geth node deployment
./scripts/docker/run.sh init

# run your local geth node for the first time
./scripts/docker/run.sh mine

# Geth will take a while to generate the DAG. Wait for this to be completed before running functional tests
# To find out how the DAG generation is going check
docker logs geth-node -f

# You can, however, already run unit/integration tests
./scripts/tests/run_unit_tests.sh
./scripts/tests/run_integration_tests.sh
```
The DAG generation progress will take up to 45 minutes. When you see `Commit new mining work` for the first time, the sync is done.

## Running Tests

Install packages and dependencies
```bash
dep ensure
```
Run only unit tests
```bash
./scripts/tests/run_unit_tests.sh
```

Run only integration Tests:
```bash
./scripts/tests/run_integration_tests.sh
```

To run functional tests a few other components need to be set up.
- Geth node needs to be up and running
- Contracts need to be deployed
    - Run contract migration (fetched by ENV_VAR CENT_ETHEREUM_CONTRACTS_DIR under `scripts/test-dependencies/test-ethereum/env_vars.sh` )
- Local account keys need to be set and able to call the right contracts on Ethereum

To do this setup + run all the tests (unit, integration, functional) use the `test_wrapper.sh` script.

If you are running this for the first time, make sure to have run
```bash
./scripts/docker/run.sh init
./scripts/docker/run.sh mine
```
beforehand and made sure that the DAG generation was completed as described in the setup chapter above.

Then run the whole test-suite with
```bash
./scripts/test_wrapper.sh
```

### Troubleshooting functional test setup

One of the most-likely issues during your first run of the `./scripts/test_wrapper.sh` will be that your geth node has not yet synced (if run against rinkeby) or finished building the DAG (if running locally).
This issue will likely show up with an error about gas limits. That's a misleading error by Truffle. Check on your DAG generation status via `docker logs geth-node -f`. If the DAG is still generating, wait.

Another case (if you ran your geth node manually via `./scripts/docker/run.sh` - instead of using `./scripts/test_wrapper.sh` - make sure to run it with the `mine` parameter instead of `local`. Otherwise it will not mine new transactions.


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


## Run a Geth node locally or Rinkeby environments

We make use of Docker Compose locally as it is easy and clear to bundle volume and environment configurations:
Docker Compose files live in `./scripts/docker`

#### Run as local node with mining enabled
The first time, initialize the config" `./scripts/docker/run.sh init`

Then run the local node via `./scripts/docker/run.sh local`
By default it uses:
- ETH_DATADIR=${HOME}/Library/Ethereum
- RPC_PORT=9545
- WS_PORT=9546


#### Run local peer connected to Rinkeby
Let's run the rinkeby local node:
`./scripts/docker/run.sh rinkeby`
By default it uses:
- ETH_DATADIR=${HOME}/Library/Ethereum
- RPC_PORT=9545

Override those when needed

Let it catch up for a while until is fully synced with the remote peer

#### Checking on your local geth node
To see what's up with your local geth node (e.g. to see how the DAG generation is going or if it is mining) use
```bash
docker logs geth-node -f
```

#### Attaching to your local geth node
In order to attach via geth to this node running in docker run
```bash
geth attach ws://localhost:9546
```

## Run Integration Tests against Local/Integration/Rinkeby Environments

### Configure local mining + run integration/functional tests
  - Remove running container if any:
    - docker rm -f geth-node
  - Clear up ~/Library/Ethereum/8383 folder (keep in mind this will clear up all previous data you had before)
    - cd ~/Library/Ethereum/8383
    - rm -Rf files/ geth/ geth.ipc keystore/ .ethash/
  - In go-centrifuge project run:
    - ./scripts/docker/run.sh init
    - ./scripts/docker/run.sh mine
    - Wait until DAG is generated:
      - docker logs geth-node 2>&1 | grep 'mined potential block'
  - Run contract migration to generate local contract address artifact:
    - In centrifuge-ethereum-contracts project:
      - ./scripts/migrate.sh local
      - Verify that ./deployments/local.json has been generated (note that local.json is excluded in .gitignore)
  - Run tests:
    - To run only integration tests:
      - ./scripts/tests/run_integration_tests.sh

### Configure node to point to integration + run integration/functional tests
  - Remove running container if any:
    - docker rm -f geth-node
  - Clear up ~/Library/Ethereum/8383 folder (keep in mind this will clear up all previous data you had before) (no need if node has synced before with peer)
    - cd ~/Library/Ethereum/8383
    - rm -Rf files/ geth/ geth.ipc keystore/ .ethash/
  - In go-centrifuge project run:
    - ./scripts/docker/run.sh init (no need if node has synced before with peer)
    - /.scripts/docker/run.sh local
    - Wait until node is in sync with remote peer:
      - geth attach ws://localhost:9546 --exec "net.peerCount" > 0
      - geth attach ws://localhost:9546 --exec "eth.syncing" -> false
  - Run tests:
    - To run only integration tests (3-4 mins):
      - CENT_CENTRIFUGENETWORK='centrifugeRussianhillEthIntegration' TEST_TARGET_ENVIRONMENT='integration' ./scripts/tests/run_integration_tests.sh

### Configure node to point to rinkeby + run integration/functional tests
  - Remove running container if any:
    - docker rm -f geth-node
  - In go-centrifuge project run:
    - ./scripts/docker/run.sh rinkeby
    - Wait until node is in sync with remote peer (1-2 hours):
      - geth attach ws://localhost:9546 --exec "net.peerCount" > 0 (rinkeby takes additional time to sync as it needs a peer to pull from, and has shortage of full node peers)
      - geth attach ws://localhost:9546 --exec "eth.syncing" -> false
  - Run tests:
    - To run only integration tests:
      - CENT_CENTRIFUGENETWORK='centrifugeRussianhillEthRinkeby' TEST_TARGET_ENVIRONMENT='rinkeby' CENT_ETHEREUM_ACCOUNTS_MAIN_KEY='$JSON_KEY' CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD="$PASS" CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS="$ADDR" ./scripts/tests/run_integration_tests.sh

### Configure node to point to infura-rinkeby + run integration/functional tests ####
  - Run tests:
    - To run only integration tests:
      - CENT_ETHEREUM_TXPOOLACCESSENABLED=false CENT_ETHEREUM_NODEURL='wss://rinkeby.infura.io/ws/MtCWERMbJtnqPKI8co84' CENT_CENTRIFUGENETWORK='centrifugeRussianhillEthRinkeby' TEST_TARGET_ENVIRONMENT='rinkeby' CENT_ETHEREUM_ACCOUNTS_MAIN_KEY='$JSON_KEY' CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD="$PASS" CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS="$ADDR" ./scripts/tests/run_integration_tests.sh

## Ethereum Contract Bindings

To create the go bindings for the deployed truffle contract, use the following command:

```bash
abigen --abi abi/AnchorRegistry.abi --pkg anchor --type EthereumAnchorRegistryContract --out ${GOPATH}/src/github.com/centrifuge/go-centrifuge/centrifuge/anchor/ethereum_anchor_registry_contract.go
```

and then copy the `ethereum_anchor_registry_contract.go` file to `centrifuge/anchor/`. You will also need to modify the file to add the following imports:

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
make gen_proto
```
