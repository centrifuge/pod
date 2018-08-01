#!/usr/bin/env bash

GATE_IP=`kubectl get svc -n spinnaker spin-gate -o jsonpath='{.status.loadBalancer.ingress[0].ip}'`
DECK_IP=`kubectl get svc -n spinnaker spin-deck -o jsonpath='{.status.loadBalancer.ingress[0].ip}'`

hal config security api edit --override-base-url http://${GATE_IP}:8084
hal config security ui edit --override-base-url http://${DECK_IP}:9000