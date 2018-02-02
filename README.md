# Centrifuge Protocol POC
Project Structure taken from: https://github.com/golang-standards/project-layout and https://github.com/ethereum/go-ethereum

Quick Start:
- mkdir -p $GOPATH/src/github.com/CentrifugeInc/go-centrifuge/
- git clone git@github.com:CentrifugeInc/go-centrifuge.git $GOPATH/src/github.com/CentrifugeInc/go-centrifuge
- go get -u github.com/kardianos/govendor
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
