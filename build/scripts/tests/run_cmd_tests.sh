#!/usr/bin/env bash

RED='\033[0;31m'
NC='\033[0m'

# Generate centrifuge binary if not in travis (travis would have already generated it)
if [[ "X${TRAVIS}" != "Xtrue" ]]; then
  make install
fi

CENTBIN=$GOPATH/bin/centrifuge
CENTCFG=$HOME/datadir

# Version check
vout=`$CENTBIN version`
status=$?
if [ $status -ne 0 ]; then
  echo "${RED}Error [version]:${NC}"
  echo "${vout}"
  exit $status
fi

export CENT_ETHEREUM_CONTRACTS_DIR=vendor/github.com/centrifuge/centrifuge-ethereum-contracts
export CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYFACTORY=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/IdentityFactory.json | jq -r --arg NETWORK_ID "8383" '.networks[$NETWORK_ID].address' | tr -d '\n'`
export CENT_NETWORKS_TESTING_CONTRACTADDRESSES_IDENTITYREGISTRY=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/IdentityRegistry.json | jq -r --arg NETWORK_ID "8383" '.networks[$NETWORK_ID].address' | tr -d '\n'`
export CENT_NETWORKS_TESTING_CONTRACTADDRESSES_ANCHORREPOSITORY=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/AnchorRepository.json | jq -r --arg NETWORK_ID "8383" '.networks[$NETWORK_ID].address' | tr -d '\n'`
export CENT_NETWORKS_TESTING_CONTRACTADDRESSES_PAYMENTOBLIGATION=`cat $CENT_ETHEREUM_CONTRACTS_DIR/build/contracts/PaymentObligation.json | jq -r --arg NETWORK_ID "8383" '.networks[$NETWORK_ID].address' | tr -d '\n'`

# Create config check
cfgout=`$CENTBIN createconfig -n testing -t $CENTCFG -z build/scripts/test-dependencies/test-ethereum/migrateAccount.json`
status=$?
if [ $status -ne 0 ]; then
  echo -e "${RED}Error [createconfig]:${NC}"
  echo "${cfgout}"
else
  if [ ! -f $CENTCFG/config.yaml ]; then
    echo "Config file not found!"
    status=1
  fi
fi

exit $status
