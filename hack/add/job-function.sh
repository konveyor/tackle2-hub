#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/controls/job-function -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "role": "Administrator"
}' | jq -M .
