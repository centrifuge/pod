#!/usr/bin/env bash

my_dir="$(dirname "$0")"
source "${my_dir}/env_vars.sh"

geth attach $DATA_DIR/geth.ipc
