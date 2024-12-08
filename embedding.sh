#!/bin/bash

if [ -f .env ]; then
  source .env
else
  echo ".env file not found. Exiting."
  exit 1
fi

if [ -z "$EMBED_PORT" ]; then
  echo "invalid env, check if set: EMBED_PORT"
  exit 1
fi

model=intfloat/multilingual-e5-large-instruct
volume=$PWD/volumes

res=$(docker ps | grep $EMBED_CONTAINER_NAME | grep healthy | wc -l)

if [ "$res" -eq 1 ]; then
  echo "$EMBED_CONTAINER_NAME is already running and healthy."
  exit 0
fi

if docker ps -a | grep -q $EMBED_CONTAINER_NAME; then
  echo "Container $EMBED_CONTAINER_NAME exists but is not running. Starting it."
  docker start $EMBED_CONTAINER_NAME
  exit 0
fi

echo "Starting $EMBED_CONTAINER_NAME..."
docker run -d --name $EMBED_CONTAINER_NAME --gpus all -p $EMBED_PORT:80 -v $volume:/data --pull always ghcr.io/huggingface/text-embeddings-inference:1.5 --model-id $model


# use: (return a 1024 length list of numbers for each input)
#    curl 127.0.0.1:5000/embed \
#        -X POST \
#        -d '{"inputs": ["What is Deep Learning?", "It's hot today."]}' \
#        -H 'Content-Type: application/json'
#
# documentation
# https://huggingface.co/docs/text-embeddings-inference/en/quick_tour
