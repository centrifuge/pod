#!/usr/bin/env bash

echo "bridge node was running? [${BRIDGE_DOCKER_CONTAINER_WAS_RUNNING}]"
if [ -n "${BRIDGE_DOCKER_CONTAINER_WAS_RUNNING}" ]; then
    echo "Container ${BRIDGE_DOCKER_CONTAINER_NAME} is already running. Not starting again."
    exit 0;
else
    echo "Container ${BRIDGE_DOCKER_CONTAINER_NAME} is not currently running. Going to start."
fi

# Setup
PARENT_DIR=$(pwd)

################## Run Bridge #########################
"${PARENT_DIR}"/build/scripts/docker/run.sh bridge

echo "Waiting for Bridge to Start Up ..."
maxCount=100
echo "MaxCount: $maxCount"
count=0
while true
do
  started=$(docker logs bridge 2>&1 | grep 'Block not')
  if [ "$started" != "" ]; then
    echo "Bridge successfully started"
    break
  elif [ $count -ge $maxCount ]; then
    echo "Timeout Starting out Bridge, printing logs:"
    cat /tmp/bridge-0.log
    exit 1
  elif [ "$(docker logs bridge 2>&1 | grep 'no bytecode found at')" != "" ]; then
    cat /tmp/bridge-0.log
    echo "Please force run migrations again using 'FORCE_MIGRATE=true'"
    exit 1
  fi
  sleep 2;
  ((count++))
done
