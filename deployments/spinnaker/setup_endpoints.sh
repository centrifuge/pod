#!/usr/bin/env bash

kubectl patch svc spin-gate -p '{"spec":{"type":"LoadBalancer"}}' -n spinnaker
kubectl patch svc spin-deck -p '{"spec":{"type":"LoadBalancer"}}' -n spinnaker
echo "Wait for until endpoints populated"