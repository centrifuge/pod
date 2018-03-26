#!/bin/bash

export CENT_ETHEREUM_CONTEXTWAITTIMEOUT="120s"
export CENT_ETHEREUM_GETHIPC=$HOME/.centrifuge/geth_test_network/geth.ipc
export CENT_ETHEREUM_GASLIMIT=4712388
export CENT_ETHEREUM_GASPRICE=40000
export CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD=ZhXfpAc#vHu4JTELA
export CENT_ETHEREUM_ACCOUNTS_MAIN_KEY='{"address":"838f7dca284eb69a9c489fe09c31cff37defdeca","crypto":{"cipher":"aes-128-ctr","ciphertext":"b16312912c00712f02b43ed3cdd3b3172195329415527f7ee218656888aa5d92","cipherparams":{"iv":"19494c514fae0e4d83d9a7e464e89e29"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"e9b7cf9b55eab4a54f6f6f5af98ca6add2ca49147d37f99a5fa26a89e9003517"},"mac":"04805d48727a24cc3ee2ac2198f7fd5be269e52ff105c125cd10b614ce0d856d"},"id":"cd3800bc-c85d-457b-925b-09d809d6b06e","version":3}'
export CENT_ANCHOR_ETHEREUM_ANCHORREGISTRYADDRESS="0xd1f76467ca9931c2a9a80decf94193b76adeffcf"

go test ./centrifuge/anchor/anchor_registry_integration_test.go --ethereum -v