#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/proxies -d \
'{
    "kind": "http",
    "host":"myhost",
    "port": 80,
    "identity": {
      "id": 1
    }
}' | jq -M .

curl -X POST ${host}/proxies -d \
'{
    "kind": "https",
    "host":"myhost",
    "port": 443,
    "excluded": [
      "redhat.com"
    ]
}' | jq -M .
