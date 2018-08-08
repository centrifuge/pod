#!/usr/bin/env bash

# TODO This script should be removed once we have proper functional tests

RAND_ID=$(openssl rand -base64 12)

curl -X POST "https://127.0.0.1:8082/invoice/anchor" \
-H "accept: application/json"      -H "Content-Type: application/json" \
-d "{ \"document\": { \"coreDocument\": {\"documentIdentifier\": \"$RAND_ID\"},\"documentIdentifier\": \"$RAND_ID\", \"data\": {\"recipient_country\":\"US\", \"currency\": \"USD\", \"net_amount\": \"1501\"}}}" \
-k