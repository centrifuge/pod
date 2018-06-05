#!/usr/bin/env bash
echo $DOCKERHUBPWD | wc
echo $DOCKERHUBUSER | wc
echo "$DOCKERHUBPWD" | docker login -u "$DOCKERHUBUSER" --password-stdin
