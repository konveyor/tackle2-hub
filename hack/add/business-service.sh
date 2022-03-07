#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/businessservices -d \
'{
    "createUser": "tackle",
    "name": "Marketing",
    "Description": "Marketing Dept.",
    "owner": {
      "id": 1
    }
}' | jq -M .
