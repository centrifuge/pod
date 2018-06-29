#!/usr/bin/env bash

git_sha="$(git rev-parse --short HEAD)"
docker build -t ${IMAGE_NAME}:${git_sha} .

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
docker tag "${IMAGE_NAME}:${git_sha}" "${IMAGE_NAME}:latest"
docker push ${IMAGE_NAME}:latest
docker push ${IMAGE_NAME}:${git_sha}