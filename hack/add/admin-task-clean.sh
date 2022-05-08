#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/tasks -d \
'{
    "name":"Test-mount-clean",
    "state": "Ready",
    "variant":"mount:clean",
    "priority": 1,
    "policy": "isolated",
    "addon": "admin",
    "data": {
      "path": "m2"
    }
}' | jq -M .

