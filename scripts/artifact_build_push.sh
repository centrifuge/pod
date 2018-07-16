#!/usr/bin/env bash

set -e

git_sha="$(git rev-parse --short HEAD)"
timestamp=`date -u +%Y%m%d`
tag="${timestamp}-${git_sha}"

echo "Building Alpine Docker Image"
docker build -t ${IMAGE_NAME}:${tag} .

echo "Building Ubuntu Docker Image for Linux Binary creation"
docker build -t ${IMAGE_NAME}:${tag} -f Dockerfile.ubuntu .

mkdir -p build/linux
cid=$(docker create ${IMAGE_NAME}:${tag})
docker cp $cid:/root/centrifuge build/linux/
docker rm -v $cid

echo "Creating Tar file of binary"
tar -zcvf cent-api-${tag}.tar.gz -C build/linux .

echo "Pushing Tar Artifact to GCS"
gcloud auth activate-service-account --key-file peak-vista-185616-9f70002df7eb.json
gsutil cp cent-api-${tag}.tar.gz gs://centrifuge-artifact-releases/

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker tag "${IMAGE_NAME}:${tag}" "${IMAGE_NAME}:latest"
docker push ${IMAGE_NAME}:latest
docker push ${IMAGE_NAME}:${tag}