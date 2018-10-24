#!/usr/bin/env bash

# Get latest Anchor and Identity Registry Addresses from contract json
export TEST_TIMEOUT=${TEST_TIMEOUT:-600s}
export TEST_TARGET_ENVIRONMENT=${TEST_TARGET_ENVIRONMENT:-'localgeth'}
export CENT_CENTRIFUGENETWORK=${CENT_CENTRIFUGENETWORK:-'testing'}

## Making Env Var Name dynamic
cent_upper_network=`echo $CENT_CENTRIFUGENETWORK | awk '{print toupper($0)}'`
tempIdentityFactory="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_IDENTITYFACTORY"
printf -v $tempIdentityFactory `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.IdentityFactory.address' | tr -d '\n'`
tempIdentityRegistry="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_IDENTITYREGISTRY"
printf -v $tempIdentityRegistry `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.IdentityRegistry.address' | tr -d '\n'`
tempAnchorRepository="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_ANCHORREPOSITORY"
printf -v $tempAnchorRepository `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.AnchorRepository.address' | tr -d '\n'`
tempPaymentObligation="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_PAYMENTOBLIGATION"
printf -v $tempPaymentObligation `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.PaymentObligation.address' | tr -d '\n'`

export $tempIdentityFactory
export $tempIdentityRegistry
export $tempAnchorRepository
export $tempPaymentObligation

vtempIdentityFactory=$(eval "echo \"\$$tempIdentityFactory\"")
vtempIdentityRegistry=$(eval "echo \"\$$tempIdentityRegistry\"")
vtempAnchorRepository=$(eval "echo \"\$$tempAnchorRepository\"")
vtempPaymentObligation=$(eval "echo \"\$$tempPaymentObligation\"")
#

echo "ANCHOR REPOSITORY ADDRESS: ${vtempAnchorRepository}"
echo "IDENTITY REGISTRY ADDRESS: ${vtempIdentityRegistry}"
echo "IDENTITY FACTORY ADDRESS: ${vtempIdentityFactory}"
echo "PAYMENT OBLIGATION ADDRESS: ${vtempPaymentObligation}"


if [ -z $vtempIdentityFactory ] || [ -z $vtempIdentityRegistry ] || [ -z $vtempAnchorRepository ] || [ -z $vtempPaymentObligation ]; then
    echo "One of the required contract addresses is not set. Aborting."
    exit -1
fi