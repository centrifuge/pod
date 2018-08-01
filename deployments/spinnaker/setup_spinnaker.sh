#!/usr/bin/env bash

set -e

PROJECT=${1:-peak-vista-185616}
COMPUTE_ZONE=${2:-us-central1-a}
SPINNAKER_CLUSTER_ID=${3:-centrifuge-spinnaker-test}

echo "Setting up project id [$PROJECT]"
gcloud config set project $PROJECT
echo "Setting up COMPUTE ZONE [$COMPUTE_ZONE]"
gcloud config set compute/zone $COMPUTE_ZONE
echo "Ensuring kubectl in installed"
sudo apt-get install kubectl
echo "Downloading k8s cluster credentials for cluster [$SPINNAKER_CLUSTER_ID]"
gcloud container clusters get-credentials $SPINNAKER_CLUSTER_ID
echo "Installing latest Halyard version"
curl -O https://raw.githubusercontent.com/spinnaker/halyard/master/install/debian/InstallHalyard.sh
sudo bash InstallHalyard.sh -y
set +e
echo "Creating spinnaker namespace"
ns_exists=`kubectl get namespaces | grep spinnaker`
if [ $? -ne 0 ]; then
	kubectl create namespaces spinnaker
fi
CONTEXT=$(kubectl config current-context)
echo "Configuring Spinnaker Service Account + Cluster Roles"
kubectl apply --context $CONTEXT -f ssa.yaml
SECRET_NAME=$(kubectl get sa spinnaker-service-account --context $CONTEXT -n spinnaker -o jsonpath='{.secrets[0].name}')
TOKEN=$(kubectl get secret --context $CONTEXT $SECRET_NAME -n spinnaker -o jsonpath='{.data.token}' | base64 --decode)
kubectl config set-credentials ${CONTEXT}-token-user --token $TOKEN
kubectl config set-context $CONTEXT --user ${CONTEXT}-token-user

echo "Configuring Halyard - Docker Registry"
ADDRESS=index.docker.io
REPOSITORIES=centrifugeio/go-centrifuge
USERNAME=mikiquantum
hal config provider docker-registry enable
hal config provider docker-registry account add dockerhub --address $ADDRESS     --repositories $REPOSITORIES     --username $USERNAME     --password

echo "Configuring Halyard - Distributed Kubernetes"
hal config version edit --version 1.8.4
hal config provider kubernetes enable
hal config provider kubernetes account add spinnaker --provider-version v2 --context $CONTEXT --docker-registries dockerhub
hal config features edit --artifacts true
hal config deploy edit --type distributed --account-name spinnaker

echo "Configuring Halyard - GCS"
PROJECT=$(gcloud info --format='value(config.project)')
BUCKET_LOCATION=us
SERVICE_ACCOUNT_DEST=~/.gcp/gcs-account.json
hal config storage gcs edit --project $PROJECT --bucket-location $BUCKET_LOCATION --json-path $SERVICE_ACCOUNT_DEST
hal config storage edit --type gcs

echo "Backing up Halyard configuration"
hal backup create