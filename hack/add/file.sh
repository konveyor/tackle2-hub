#!/bin/bash

host="${HOST:-localhost:8080}"
path="${1:-/etc/hosts}"
name=$(basename ${path})

curl -F 'file=@/etc/hosts' http://${host}/files/${name} | jq .

