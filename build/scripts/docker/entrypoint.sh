#!/usr/bin/env sh

set -x

CENT_MODE=${CENT_MODE:-run}

ETHKEY=`cat /root/.centrifuge/config/eth.key`
ETHPWD=`cat /root/.centrifuge/config/eth.pwd`

ETHEREUM_ACCOUNTS_MAIN_KEY=$ETHKEY ETHEREUM_ACCOUNTS_MAIN_PASSWORD=$ETHPWD /root/centrifuge ${CENT_MODE} --config /root/.centrifuge/config/config.yaml $@
