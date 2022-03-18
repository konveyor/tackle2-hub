#!/bin/bash

host="${HOST:-localhost:8080}"

# ID to update (default:1)
id="${1:-1}"


curl -X PUT ${host}/tasks/${id}/submit -d \
'{
    "name":"Test",
    "locator": "app.1.test",
    "addon": "test",
    "application": {"id": 1},
    "data": {
      "path": "/etc"
    }
}'
