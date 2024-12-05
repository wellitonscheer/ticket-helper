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

docker run -d -p $ATTU_PORT:3000 -e MILVUS_URL=$MY_IP:$MILVUS_PORT zilliz/attu:v2.4
