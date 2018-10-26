#!/usr/bin/env sh

set -x

CENT_MODE=${CENT_MODE:-run}

/root/go-centrifuge ${CENT_MODE} --config /root/.centrifuge/config/config.yaml $@
