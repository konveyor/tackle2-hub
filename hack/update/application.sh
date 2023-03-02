#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/applications/1 -d \
'{
    "name":"Dog-updated",
    "description": "Dog application.-updated",
    "businessService": {"id":1},
    "identities": [
      {"id":1}
    ],
    "facts": {
       "A":"1-updated",
       "B":"2-updated",
       "C":"3"
    },
    "tags":[
      {"id":1},
      {"id":2},
      {"id":3}
    ]
}' | jq -M .
