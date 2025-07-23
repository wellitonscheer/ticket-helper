#!/bin/bash

if [ -f .env ]; then
  source .env
else
  echo ".env file not found. Exiting."
  exit 1
fi

if [ -z "$POSTGRES_CONTAINER_NAME" ]; then
  echo "invalid env, check if set: POSTGRES_CONTAINER_NAME"
  exit 1
fi

if [ -z "$POSTGRES_USER" ]; then
  echo "invalid env, check if set: POSTGRES_USER"
  exit 1
fi

if [ -z "$POSTGRES_DB" ]; then
  echo "invalid env, check if set: POSTGRES_DB"
  exit 1
fi

if [ -z "$POSTGRES_PASSWORD" ]; then
  echo "invalid env, check if set: POSTGRES_PASSWORD"
  exit 1
fi

volume=$PWD/volumes/postgres

res=$(docker ps | grep $POSTGRES_CONTAINER_NAME | grep healthy | wc -l)

if [ "$res" -eq 1 ]; then
  echo "$POSTGRES_CONTAINER_NAME is already running and healthy."
  exit 0
fi

if docker ps -a | grep -q $POSTGRES_CONTAINER_NAME; then
  echo "Container $POSTGRES_CONTAINER_NAME exists but is not running. Starting it."
  docker start $POSTGRES_CONTAINER_NAME
  exit 0
else
    echo "Starting $POSTGRES_CONTAINER_NAME..."
    docker run -d \
        --name $POSTGRES_CONTAINER_NAME \
        -e POSTGRES_USER=$POSTGRES_USER \
        -e POSTGRES_DB=$POSTGRES_DB \
        -e POSTGRES_PASSWORD=$POSTGRES_PASSWORD \
        -v $volume:/var/lib/postgresql/data \
        -p 5432:5432 \
        --pull always pgvector/pgvector:pg17
fi
