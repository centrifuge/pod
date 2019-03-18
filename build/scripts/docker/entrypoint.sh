#!/usr/bin/env sh

set -x

CENT_MODE=${CENT_MODE:-run}

ETHKEY=`cat /root/.centrifuge/secrets/eth.key`
ETHPWD=`cat /root/.centrifuge/secrets/eth.pwd`

CENT_ETHEREUM_ACCOUNTS_MAIN_KEY=$ETHKEY CENT_ETHEREUM_ACCOUNTS_MAIN_PASSWORD=$ETHPWD /root/centrifuge ${CENT_MODE} --config /root/.centrifuge/config/config.yaml $@
