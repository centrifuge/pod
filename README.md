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
curl https://glide.sh/get | sh
cd $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
glide update
```

Make sure you have docker-compose installed, usually comes bundled with Mac OS Docker. Otherwise: https://docs.docker.com/compose/install/ 

Build, test & run
-----------------

Build/install:
```
cd $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
go install ./centrifuge/
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

and then run (only for unit tests):

```
reflex -r centrifuge/ go test ./... -tags=unit
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
