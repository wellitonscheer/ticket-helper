#!/bin/bash

model=intfloat/multilingual-e5-large-instruct
volume=$PWD/volumes

docker run -d --gpus all -p 5000:80 -v $volume:/data --pull always ghcr.io/huggingface/text-embeddings-inference:1.5 --model-id $model

# use: (return a 1024 length list of numbers for each input)
#    curl 127.0.0.1:5000/embed \
#        -X POST \
#        -d '{"inputs": ["What is Deep Learning?", "It's hot today."]}' \
#        -H 'Content-Type: application/json'
#
# documentation
# https://huggingface.co/docs/text-embeddings-inference/en/quick_tour