# Centrifuge Protocol POC
Project Structure taken from: https://github.com/golang-standards/project-layout and https://github.com/ethereum/go-ethereum

Quick Start:
- mkdir -p $GOPATH/src/github.com/lucasvo/cent-app-playground
- git clone git@github.com:lucasvo/cent-app-playground.git $GOPATH/src/github.com/lucasvo/cent-app-playground
- go get -u github.com/kardianos/govendor
- cd $GOPATH/src/github.com/lucasvo/cent-app-playground
- govendor sync

Build pkg and bin:
- cd $GOPATH/src/github.com/lucasvo/cent-app-playground
- govendor build ./centrifuge || go build ./centrifuge
- govendor install ./centrifuge || go install ./centrifuge

Run Tests:
- govendor test ./centrifuge || go test ./centrifuge
- reflex -r centrifuge/ govendor test ./centrifuge

Run Cent-Constellation Nodes:
- Requires to have constellation-node installed and available in the PATH, follow instructions here: https://github.com/jpmorganchase/constellation/blob/master/README.md
- mkdir -p /tmp/node1/data && constellation-node --generatekeys=/tmp/node1/data/node1
- mkdir -p /tmp/node2/data && constellation-node --generatekeys=/tmp/node2/data/node2
- mkdir -p /tmp/node3/data && constellation-node --generatekeys=/tmp/node3/data/node3
- govendor install ./cmd/cent from project folder
- Make sure $GOPATH/bin is on your PATH
- Modify /etc/hosts by adding:
```
127.0.0.1 node1
127.0.0.1 node2
127.0.0.1 node3
```
- In terminal one do: `cent run-node /tmp/node1/data/node1.pub /tmp/node1/data/constellation.ipc $GOPATH/src/github.com/lucasvo/cent-app-playground/resources/node1.conf 8000`
- In terminal two do: `cent run-node /tmp/node2/data/node2.pub /tmp/node2/data/constellation.ipc $GOPATH/src/github.com/lucasvo/cent-app-playground/resources/node2.conf 8001`
- In terminal three do: `cent run-node /tmp/node3/data/node3.pub /tmp/node3/data/constellation.ipc $GOPATH/src/github.com/lucasvo/cent-app-playground/resources/node3.confcent run-node /tmp/node3/data/node3.pub /tmp/node3/data/constellation.ipc $GOPATH/src/github.com/lucasvo/cent-app-playground/resources/node3.conf  8002`
