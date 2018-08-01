#!/usr/bin/env bash
set -e

ACCOUNT_NAME=${1:-centrifuge-test}

echo "Downloading k8s cluster credentials for cluster [${ACCOUNT_NAME}]"
gcloud container clusters get-credentials $ACCOUNT_NAME

CONTEXT=$(kubectl config current-context)
kubectl apply --context $CONTEXT -f ssa_kube.yaml
SECRET_NAME=$(kubectl get sa spinnakera-service-account --context $CONTEXT -n default -o jsonpath='{.secrets[0].name}')
TOKEN=$(kubectl get secret --context $CONTEXT $SECRET_NAME -n default -o jsonpath='{.data.token}' | base64 --decode)
kubectl config set-credentials ${CONTEXT}-token-user --token $TOKEN
kubectl config set-context $CONTEXT --user ${CONTEXT}-token-user

hal config provider kubernetes account add $ACCOUNT_NAME --provider-version v2 --context $CONTEXT --docker-registries dockerhub