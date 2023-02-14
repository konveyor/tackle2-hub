#!/bin/bash

host="${HOST:-localhost:8080}"
path="${1:-/etc/hosts}"
name=$(basename ${path})

curl -F "file=@${path}" http://${host}/files/${name} | jq .

