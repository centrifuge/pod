# Centrifuge Protocol POC
Project Structure taken from: https://github.com/golang-standards/project-layout and https://github.com/ethereum/go-ethereum

Quick Start:
- mkdir -p $GOPATH/src/github.com/CentrifugeInc/go-centrifuge/
- git clone git@github.com:CentrifugeInc/go-centrifuge.git $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
- go get -u github.com/kardianos/govendor
- go get github.com/ethereum/go-ethereum
- go get github.com/mitchellh/go-homedir
- go get github.com/spf13/cobra
- go get github.com/spf13/viper
- go get github.com/CentrifugeInc/centrifuge-ethereum-contracts/centrifuge/witness
- cd $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
- govendor sync

Build pkg and bin:
- cd $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
- govendor build +p && govendor install +p

Run Tests:
- govendor test ./centrifuge || go test ./centrifuge/*
- reflex -r centrifuge/ govendor test ./centrifuge/*

Run Cent-Constellation Nodes:
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
==========================
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


