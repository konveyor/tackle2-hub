#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/tasks -d \
'{
    "name":"Test-mount-report",
    "state": "Ready",
    "variant":"mount:report",
    "priority": 1,
    "addon": "admin",
    "data": {
      "path": "m2"
    }
}' | jq -M .

