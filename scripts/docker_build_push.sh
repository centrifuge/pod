#!/usr/bin/env bash

echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
git_sha="$(git rev-parse --short HEAD)"
docker tag "$IMAGE_NAME" "${IMAGE_NAME}:latest"
docker tag "$IMAGE_NAME" "${IMAGE_NAME}:${git_sha}"
docker push ${IMAGE_NAME}:latest
docker push ${IMAGE_NAME}:${git_sha}