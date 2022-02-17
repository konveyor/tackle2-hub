#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/controls/business-service -d \
'{
    "createUser": "tackle",
    "name": "Marketing",
    "Description": "Marketing Dept.",
    "owner": {
      "id": 1
    }
}' | jq -M .
