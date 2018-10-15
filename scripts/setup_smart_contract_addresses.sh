#!/usr/bin/env bash

# Get latest Anchor and Identity Registry Addresses from contract json
export TEST_TIMEOUT=${TEST_TIMEOUT:-600s}
export TEST_TARGET_ENVIRONMENT=${TEST_TARGET_ENVIRONMENT:-'local'}
export CENT_CENTRIFUGENETWORK=${CENT_CENTRIFUGENETWORK:-'testing'}

## Making Env Var Name dynamic
cent_upper_network=`echo $CENT_CENTRIFUGENETWORK | awk '{print toupper($0)}'`
tempAnchorRegistry="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_ANCHORREGISTRY"
printf -v $tempAnchorRegistry `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.AnchorRegistry.address' | tr -d '\n'`
tempIdentityFactory="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_IDENTITYFACTORY"
printf -v $tempIdentityFactory `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.IdentityFactory.address' | tr -d '\n'`
tempIdentityRegistry="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_IDENTITYREGISTRY"
printf -v $tempIdentityRegistry `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.IdentityRegistry.address' | tr -d '\n'`
tempIdentityRepository="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_ANCHORREPOSITORY"
printf -v $tempIdentityRepository `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.AnchorRepository.address' | tr -d '\n'`
tempPaymentObligation="CENT_NETWORKS_${cent_upper_network}_CONTRACTADDRESSES_PAYMENTOBLIGATION"
printf -v $tempPaymentObligation `cat $CENT_ETHEREUM_CONTRACTS_DIR/deployments/$TEST_TARGET_ENVIRONMENT.json | jq -r '.contracts.PaymentObligation.address' | tr -d '\n'`

export $tempAnchorRegistry
export $tempIdentityFactory
export $tempIdentityRegistry
export $tempIdentityRepository
export $tempPaymentObligation

vtempAnchorRegistry=$(eval "echo \"\$$tempAnchorRegistry\"")
vtempIdentityFactory=$(eval "echo \"\$$tempIdentityFactory\"")
vtempIdentityRegistry=$(eval "echo \"\$$tempIdentityRegistry\"")
vtempIdentityRepository=$(eval "echo \"\$$tempIdentityRepository\"")
vtempPaymentObligation=$(eval "echo \"\$$tempPaymentObligation\"")
#

echo "ANCHOR REGISTRY ADDRESS: ${vtempAnchorRegistry}"
echo "ANCHOR REPOSITORY ADDRESS: ${vtempIdentityRepository}"
echo "IDENTITY REGISTRY ADDRESS: ${vtempIdentityRegistry}"
echo "IDENTITY FACTORY ADDRESS: ${vtempIdentityFactory}"
echo "PAYMENT OBLIGATION ADDRESS: ${vtempPaymentObligation}"


if [ -z $vtempAnchorRegistry ] || [ -z $vtempIdentityFactory ] || [ -z $vtempIdentityRegistry ] || [ -z $vtempIdentityRepository ]; then
    echo "One of the required contract addresses is not set. Aborting."
    exit -1
fi