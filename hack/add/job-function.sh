#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/jobfunctions -d \
'{
    "name": "tackle"
}' | jq -M .
