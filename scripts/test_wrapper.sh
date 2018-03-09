#!/bin/bash
set -a

# Setup
local_dir="$(dirname "$0")"
PARENT_DIR=`pwd`

################# Run Dependencies #########################
for path in ${local_dir}/test-dependencies/*; do
    [ -d "${path}" ] || continue # if not a directory, skip
    source "${path}/env_vars.sh" # Every dependency should have env_vars.sh + run.sh executable files
    echo "Executing [${path}/run.sh]"
    ${path}/run.sh
    if [ $? -ne 0 ]; then
      exit 1
    fi
done
############################################################

################# Prepare for tests ########################
cd $CENT_ETHEREUM_CONTRACTS_DIR
npm install

# Unlock User to Run Migration and Run it
geth attach "http://localhost:${RPC_PORT}" --exec "personal.unlockAccount('0x${CENT_ETHEREUM_ACCOUNTS_MIGRATE_ADDRESS}', '${CENT_ETHEREUM_ACCOUNTS_MIGRATE_PASSWORD}')" && truffle migrate --network localgeth -f 2
status=$?

export CENT_ANCHOR_ETHEREUM_ANCHORREGISTRYADDRESS=`cat build/contracts/AnchorRegistry.json | jq -r --arg NETWORK_ID "${NETWORK_ID}" '.networks[$NETWORK_ID].address' | tr -d '\n'`
cd ${PARENT_DIR}

echo "ANCHOR ADDRESS: ${CENT_ANCHOR_ETHEREUM_ANCHORREGISTRYADDRESS}"
############################################################

################# Run Tests ################################
# Exclude the vendor dir from test run.
# The test runner included it on travis if not explicitly excluded.
if [ $status -eq 0 ]; then
  echo "Running Unit Tests"
  go test ./... -tags=unit
  status1=$?

  echo "Running Integration Ethereum Tests against IPC [${CENT_ETHEREUM_GETHIPC}]"
  go test ./... -tags=ethereum
  status2=$?

  # Store status of tests
  status="$(( $status1 | $status2 ))"
fi
############################################################

################# CleanUp ##################################
echo "Bringing GETH Daemon Down"
killall -HUP geth
rm -Rf $DATA_DIR/geth.ipc
rm -Rf $DATA_DIR/geth.out

# Cleaning extra DAG file, so we do not cache it - travis
if [[ "X${RUN_CONTEXT}" == "Xtravis" ]];
then
  new_dag=`ls -ltr $DATA_DIR/.ethash/* | tail -1 | awk '{print $9}' | tr -d '\n'`
  rm -Rf $new_dag
fi
############################################################

################# Propagate test status ####################
exit $status
############################################################