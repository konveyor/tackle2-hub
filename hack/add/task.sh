#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/tasks -d \
'{
    "name":"Test",
    "locator": "app.1.test",
    "addon": "sleeper",
    "application": {"id": 1},
    "data": {
      "path": "/etc"
    }
}' | jq -M .

