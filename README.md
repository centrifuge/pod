Centrifuge OS Client
====================
[![Build Status](https://travis-ci.com/CentrifugeInc/go-centrifuge.svg?token=Sbf68xBZUZLMB3kGTKcX&branch=master)](https://travis-ci.com/CentrifugeInc/go-centrifuge)

Project Structure taken from: https://github.com/golang-standards/project-layout and https://github.com/ethereum/go-ethereum

Setup
-----

```bash,
brew install jq
mkdir -p $GOPATH/src/github.com/CentrifugeInc/go-centrifuge/
git clone git@github.com:CentrifugeInc/go-centrifuge.git $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
```

Make sure you have docker-compose installed, usually comes bundled with Mac OS Docker. Otherwise: https://docs.docker.com/compose/install/ 

Build, test & run
-----------------

Build/install:
```
cd $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
make install
```

Run Unit+Integration Tests:
```
./scripts/test_wrapper.sh
```
This will:
* Start up a local testnet
* Run contract migration (fetched by ENV_VAR CENT_ETHEREUM_CONTRACTS_DIR under `scripts/test-dependencies/test-ethereum/env_vars.sh` )
* Run unit + integration tests

Run only Unit Tests:
```
./scripts/tests/run_unit_tests.sh
```
Run only Integration Tests:
```
./scripts/tests/run_integration_tests.sh
```

If you want to run tests continuously when a file changes, you first need to install reflex:

```
go get github.com/cespare/reflex
```

Then run (only for unit tests). It is a good idea to exclude the `vendor` and ``.idea` folders from scanning by reflex as it hogs a lot of resources for no good reason.

```
reflex -R '(^|/)vendor/|(^|/)\\.idea/' -- go test ./centrifuge/... -tags=unit
```

Or run for specific tests only:
```
reflex -R '(^|/)vendor/|(^|/)\\.idea/' -- go test ./centrifuge/invoice/... -tags=unit
```


Run a Geth node locally against Integration or Rinkeby environments
--------------------------------------------------------------------------

We make use of Docker Compose locally as it is easy and clear to bundle volume and environment configurations:
Docker Compose files live here:
`./scripts/docker`

#### Run as Local Mining mode ####
`./scripts/test-dependencies/test-ethereum/run.sh`

#### Run local peer connected to Integration ####
First we need to initialize the Ethereum Data Dir:
`./scripts/docker/run.sh init`
Now we run the custom local node that by default points to the Integration node:
`./scripts/docker/run.sh local`
By default it uses:
* ETH_DATADIR=${HOME}/Library/Ethereum
* RPC_PORT=9545
* WS_PORT=9546

Override those when needed

Let it catch up for a while until is fully synced with the remote peer

#### Run local peer connected to Rinkeby ####
Let's run the rinkeby local node:
`./scripts/docker/run.sh rinkeby`
By default it uses:
* ETH_DATADIR=${HOME}/Library/Ethereum
* RPC_PORT=9545

Override those when needed

Let it catch up for a while until is fully synced with the remote peer

 Run Integration Tests against Local/Integration/Rinkeby Environments
-------------------------------------------------------------------------
#### Configure local mining + run integration/functional tests ####
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
      
#### Configure node to point to integration + run integration/functional tests ####
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

#### Configure node to point to rinkeby + run integration/functional tests ####
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

#### Configure node to point to infura-rinkeby + run integration/functional tests ####
  - Run tests:
    - To run only integration tests:
      - CENT_ETHEREUM_TXPOOLACCESSENABLED=false CENT_ETHEREUM_NODEURL='wss://rinkeby.infura.io/ws/MtCWERMbJtnqPKI8co84' CENT_CENTRIFUGENETWORK='centrifugeRussianhillEthRinkeby' TEST_TARGET_ENVIRONMENT='rinkeby' CENT_ETHEREUM_ACCOUNTS_MAIN_KEY='$JSON_KEY' CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD="$PASS" CENT_ETHEREUM_ACCOUNTS_MAIN_ADDRESS="$ADDR" ./scripts/tests/run_integration_tests.sh

Why you should test with a "real" Ethereum
------------------------------------------
Why you should not run `testrpc` for testing with go-ethereum clients:
* Transaction IDs are randomly generated and you can not rely on finding your own transactions based on the .Hash() function.
** https://github.com/trufflesuite/ganache-cli/issues/387
* It is not possible to send more than one transaction per testrpc start as testrpc returns the pending transaction count erroneously with leading 0s - this freaks out the hex decoding and it breaks. Essentially testrpc returns for a transaction count of 1 `0x01` whereas _real_ geth returns `0x1`

Save yourself some hassle and use a local testnet running

Start your local testnet (default port 9545):
```
./scripts/test-dependencies/test-ethereum/run.sh
```

In order to attach via geth to this node running in docker run
```
geth attach ws://localhost:9546
```


Run very simple local ethscan
-----------------------------
Follow instructions here: https://github.com/carsenk/explorer

Will need to modify `scripts/test-dependencies/test-ethereum/run.sh` to add cors flag

**Note that is a pretty simple version but can list blocks and transactions**

Modify Default Config File
--------------------------
Everytime a change needs to be done in the default config `resources/default_config.yaml` it is needed to regenerate the go config data.

It is needed to install go-bindata:
```
go get -u github.com/jteeuwen/go-bindata/...
```

Then run go generate:
```
go generate ./centrifuge/config/configuration.go
```

Ethereum Contract Bindings
--------------------------

To create the go bindings for the deployed truffle contract, use the following command:

```bash
abigen --abi abi/AnchorRegistry.abi --pkg anchor --type EthereumAnchorRegistryContract --out ${GOPATH}/src/github.com/CentrifugeInc/go-centrifuge/centrifuge/anchor/ethereum_anchor_registry_contract.go
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
