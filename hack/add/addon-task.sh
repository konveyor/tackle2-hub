#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/addons/test/tasks -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"Test",
    "locator": "app.1.test",
    "data": {
      "application": 1,
      "path": "/etc"
    }
}' | jq -M .
