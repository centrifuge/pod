#!/usr/bin/env bash

echo "bridge node was running? [${BRIDGE_DOCKER_CONTAINER_WAS_RUNNING}]"
if [ -n "${BRIDGE_DOCKER_CONTAINER_WAS_RUNNING}" ]; then
    echo "Container ${BRIDGE_DOCKER_CONTAINER_NAME} is already running. Not starting again."
    exit 0;
else
    echo "Container ${BRIDGE_DOCKER_CONTAINER_NAME} is not currently running. Going to start."
fi

# Setup
PARENT_DIR=`pwd`

################## Run Bridge #########################
${PARENT_DIR}/build/scripts/docker/run.sh bridge

echo "Waiting for Bridge to Start Up ..."
maxCount=100
echo "MaxCount: $maxCount"
count=0
while true
do
  started=`docker logs bridge 2>&1 | grep 'Started'`
  if [ "$started" != "" ]; then
    echo "Bridge successfully started"
    break
  elif [ $count -ge $maxCount ]; then
    echo "Timeout Starting out Bridge"
    exit 1
  fi
  sleep 2;
  ((count++))
done

${PARENT_DIR}/build/scripts/test-dependencies/test-xbridge/add_balance.sh
