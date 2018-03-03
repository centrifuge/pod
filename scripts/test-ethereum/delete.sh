#!/bin/bash

my_dir="$(dirname "$0")"
source "${my_dir}/env_vars.sh"

rm -rf $DATA_DIR
