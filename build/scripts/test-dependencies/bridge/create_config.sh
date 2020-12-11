#!/usr/bin/env bash

config_dir=$1
bridge_address=$2
erc721_address=$3
erc20_address=$4
generic_address=$5

echo "bridge contract addresses are ${bridge_address},${erc721_address},${erc20_address},${generic_address}"

json_config="
{
  \"chains\":[
    {
      \"name\": \"eth\",
      \"type\": \"ethereum\",
      \"id\": \"0\",
      \"endpoint\": \"ws://geth:9546\",
      \"from\": \"0x88740f7A4A2b28F9B2Edb3F88452592d8f31311c\",
      \"opts\": {
          \"gasMultiplier\":\"1.25\",
          \"bridge\":\"${bridge_address}\",
          \"erc721Handler\":\"${erc721_address}\",
          \"erc20Handler\":\"${erc20_address}\",
          \"genericHandler\":\"${generic_address}\"
      }
    },
    {
      \"name\": \"substrate\",
      \"type\": \"substrate\",
      \"id\": \"1\",
      \"endpoint\": \"ws://cc:9944\",
      \"from\": \"5GrwvaEF5zXb26Fz9rcQpDWS57CtERHpNehXCPcNoHGKutQY\",
      \"opts\": { }
    }
  ]
}
"

echo "$json_config" > "$config_dir"/config.json

echo "Created config for Bridge"
cat "$config_dir"/config.toml
