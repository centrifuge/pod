# Centrifuge OS Client

[![Build Status](https://travis-ci.com/CentrifugeInc/go-centrifuge.svg?token=Sbf68xBZUZLMB3kGTKcX&branch=master)](https://travis-ci.com/CentrifugeInc/go-centrifuge)

Project Structure taken from: https://github.com/golang-standards/project-layout and https://github.com/ethereum/go-ethereum


## Setup

```bash
# dependencies and checkout of code
brew install jq
npm install -g truffle@4.0.4
mkdir -p $GOPATH/src/github.com/CentrifugeInc/go-centrifuge/
git clone git@github.com:CentrifugeInc/go-centrifuge.git $GOPATH/src/github.com/CentrifugeInc/go-centrifuge

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

To check on the DAG generation progress (will take about 30-45 minutes):
```
docker logs geth-node -f
[...]
INFO [07-27|22:03:25] Generating DAG in progress               epoch=1 percentage=50 elapsed=7m40.893s
INFO [07-27|22:03:35] Generating DAG in progress               epoch=1 percentage=51 elapsed=7m51.045s
[...]
INFO [07-27|22:19:37] Generating DAG in progress               epoch=0 percentage=98 elapsed=6m35.368s
INFO [07-27|22:19:40] Generating DAG in progress               epoch=0 percentage=99 elapsed=6m38.744s
INFO [07-27|22:19:40] Generated ethash verification cache      epoch=0 elapsed=6m38.750s
INFO [07-27|22:19:44] Generating ethash verification cache     epoch=1 percentage=93 elapsed=3.020s
INFO [07-27|22:19:44] Generated ethash verification cache      epoch=1 elapsed=3.352s
INFO [07-27|22:19:51] Generating DAG in progress               epoch=1 percentage=0  elapsed=7.177s
INFO [07-27|22:19:58] Successfully sealed new block            number=1 hash=b5a50a…d9d2e9
INFO [07-27|22:19:58] 🔨 mined potential block                  number=1 hash=b5a50a…d9d2e9
INFO [07-27|22:19:58] Commit new mining work                   number=2 txs=0 uncles=0 elapsed=985.5µs
```

When you see `Commit new mining work` for the first time, then it is time to run the functional tests.

Make sure you have docker-compose installed, usually comes bundled with Mac OS Docker. Otherwise: https://docs.docker.com/compose/install/ 


## Running Tests

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


## Build

Build & install the Centrifuge OS Node
```bash
cd $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
make install
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


## Why you should test with a "real" Ethereum

Why you should not run `testrpc` for testing with go-ethereum clients:
- Transaction IDs are randomly generated and you can not rely on finding your own transactions based on the .Hash() function.
    - https://github.com/trufflesuite/ganache-cli/issues/387
- It is not possible to send more than one transaction per testrpc start as testrpc returns the pending transaction count erroneously with leading 0s - this freaks out the hex decoding and it breaks. Essentially testrpc returns for a transaction count of 1 `0x01` whereas _real_ geth returns `0x1`

Save yourself some hassle and use a local geth node or rinkeby


## Run very simple local ethscan

Follow instructions here: https://github.com/carsenk/explorer

Will need to modify `scripts/test-dependencies/test-ethereum/run.sh` to add cors flag

**Note that is a pretty simple version but can list blocks and transactions**


## Modify Default Config File

Everytime a change needs to be done in the default config `resources/default_config.yaml` it is needed to regenerate the go config data.

It is needed to install go-bindata:
```
go get -u github.com/jteeuwen/go-bindata/...
```

Then run go generate:
```
go generate ./centrifuge/config/configuration.go
```


## Ethereum Contract Bindings

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

## Protobufs bindings

Create any new `.proto` files in its own package under `protobufs` folder.
Generating go bindings and swagger with the following command

```bash
make gen_proto
```
