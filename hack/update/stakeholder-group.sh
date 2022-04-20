#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/stakeholdergroups/1 -d \
'{
    "name": "Big Dogs",
    "description": "Group of big dogs.",
    "stakeholders": [
      {"id":1}
    ]
}' | jq -M .
