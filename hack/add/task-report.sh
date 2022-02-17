#!/bin/bash

host="${HOST:-localhost:8080}"

# ID to update (default:1)
id="${1:-1}"

curl -X POST ${host}/tasks/${id}/report -d \
'{
    "updateUser": "tackle",
    "status": "Running",
    "total": 10,
    "completed": 0,
    "activity": "addon started."
}' | jq -M .
