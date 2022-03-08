#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/buckets -d \
'{
    "application": 1,
    "name": "created-directly"
}' | jq -M .

curl -X POST ${host}/applications/1/buckets/created-for-application -d \
'{
    "name": "created-for-application"
}' | jq -M .
