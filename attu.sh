#!/bin/bash

if [ -f .env ]; then
  source .env
else
  echo ".env file not found. Exiting."
  exit 1
fi

if [ -z "$MY_IP" ] || [ -z "$ATTU_PORT" ] || [ -z "$MILVUS_PORT" ]; then
  echo "invalid env, check if set: MY_IP, ATTU_PORT, MILVUS_PORT"
  exit 1
fi

res=$(docker ps | grep $ATTU_CONTAINER_NAME | grep healthy | wc -l)

if [ "$res" -eq 1 ]; then
  echo "$ATTU_CONTAINER_NAME is already running and healthy."
  exit 0
fi

if docker ps -a | grep -q $ATTU_CONTAINER_NAME; then
  echo "Container $ATTU_CONTAINER_NAME exists but is not running. Starting it."
  docker start $ATTU_CONTAINER_NAME
  exit 0
fi

echo "Starting $ATTU_CONTAINER_NAME..."
docker run -d --name $ATTU_CONTAINER_NAME -p $ATTU_PORT:3000 -e MILVUS_URL=$MY_IP:$MILVUS_PORT zilliz/attu:v2.4
