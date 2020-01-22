#!/usr/bin/env bash

config_dir=$1
asset_address=$2

echo "asset contract is ${asset_address}"

eth_config="[[chains]]
id = 0
endpoint = \"ws://geth:9546\"
emitter = \"0x1fA38b0EfccA4228EB9e15112D4d98B0CEe3c600\"
receiver = \"$asset_address\"
from = \"0379ac69459dfb4220f8a92e19d3316efe41646e1b0164d8c42d98c558adf2dc28\""

cent_config='[[chains]]
id = 1
endpoint = "ws://cc:9944"
emitter = "0x1fA38b0EfccA4228EB9e15112D4d98B0CEe3c600"
receiver = "0x290f41e61374c715C1127974bf08a3993afd0145"
from = "0379ac69459dfb4220f8a92e19d3316efe41646e1b0164d8c42d98c558adf2dc28"'

echo "$eth_config" > "$config_dir"/config.toml
echo "" >> "$config_dir"/config.toml
echo -n "$cent_config" >> "$config_dir"/config.toml

echo "Created config for Bridge"
cat "$config_dir"/config.toml
