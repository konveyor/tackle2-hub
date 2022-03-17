#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/applications/1 -d \
'{
    "name":"Cat",
    "description": "Cat application.",
    "businessService": {"id":1},
    "identities": [
      {"id":1},
      {"id":2}
    ],
    "tags":[
      {"id":1}
    ]
}' | jq -M .
