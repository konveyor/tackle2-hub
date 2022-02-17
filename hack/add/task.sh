#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/tasks -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"Test",
    "locator": "app.1.test",
    "addon": "test",
    "data": {
      "application": 1,
      "path": "/etc"
    }
}' | jq -M .

