#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/proxies/1 -d \
'{
    "enabled": true,
    "kind": "http",
    "host":"redhat.com",
    "port": 90
}' | jq -M .

