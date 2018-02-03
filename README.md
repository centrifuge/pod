Centrifuge Protocol POC
=======================
[![Build Status](https://travis-ci.com/CentrifugeInc/go-centrifuge.svg?token=Sbf68xBZUZLMB3kGTKcX&branch=master)](https://travis-ci.com/CentrifugeInc/go-centrifuge)

Project Structure taken from: https://github.com/golang-standards/project-layout and https://github.com/ethereum/go-ethereum

Setup
-----

```bash,
mkdir -p $GOPATH/src/github.com/CentrifugeInc/go-centrifuge/
git clone git@github.com:CentrifugeInc/go-centrifuge.git $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
brew install dep && brew update dep
go get github.com/ethereum/go-ethereum
cd $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
dep ensure
rm -rf vendor/github.com/ethereum/go-ethereum
```

Build, test & run
-----------------

Build/install:
```
cd $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
go install ./centrifuge/
```

Run Tests:

```
go test ./...
```

If you want to run tests continuously when a file changes, you first need to install reflex:

```
go get github.com/cespare/reflex
```

and then run:

```
reflex -r centrifuge/ go test ./...
```

Run Cent-Constellation Nodes
----------------------------

- Requires to have constellation-node installed and available in the PATH, follow instructions here: https://github.com/jpmorganchase/constellation/blob/master/README.md
- Install according to section above
- Make sure $GOPATH/bin is on your PATH
- `cp -r resources /tmp/centrifuge/`
- Modify /etc/hosts by adding:
```
127.0.0.1 node1
127.0.0.1 node2
127.0.0.1 node3
```
- In terminal one do: `centrifuge run --config /tmp/centrifuge/node1/centrifuge_node1.yaml`
- In terminal one do: `centrifuge run --config /tmp/centrifuge/node2/centrifuge_node2.yaml`
- In terminal one do: `centrifuge run --config /tmp/centrifuge/node3/centrifuge_node3.yaml`


Ethereum Contract Bindings
--------------------------

To create the go bindings for the deployed truffle contract, use the following command:

`abigen --abi build/contracts/Witness.abi --pkg witness --type EthereumWitness --out witness_contract.go`

and then copy the `witness_contract.go` file to `centrifuge/witness/`. You will also need to modify the file to add the following imports:

```go,
import(
   	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)
```


