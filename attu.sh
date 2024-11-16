#!/bin/bash

if [ -f .env ]; then
  source .env
else
  echo ".env file not found. Exiting."
  exit 1
fi

if [ -z "$MY_IP" ]; then
  echo "MY_IP not set in .env. Exiting."
  exit 1
fi

docker run -d -p 8000:3000 -e MILVUS_URL=$MY_IP:19530 zilliz/attu:v2.4