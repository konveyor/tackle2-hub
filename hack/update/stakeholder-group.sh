#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/stakeholdergroups -d \
'{
    "name": "Big Dogs",
    "description": "Group of big dogs."
}' | jq -M .
