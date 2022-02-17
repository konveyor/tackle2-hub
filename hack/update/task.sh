#!/bin/bash

host="${HOST:-localhost:8080}"

# ID to update (default:1)
id="${1:-1}"


curl -X PUT ${host}/tasks/${id}/report -d \
'{
    "updateUser": "tackle",
    "status": "Running",
    "total": 10,
    "completed": 9,
    "activity": "reading /files/application/dog.java."
}'
