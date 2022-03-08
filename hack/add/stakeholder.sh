#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/stakeholders -d \
'{
    "name": "tackle",
    "displayName":"Elmer",
    "email": "tackle@konveyor.org",
    "role": "Administrator",
    "stakeholderGroups": [{"id": 1}],
    "jobFunction" : {"id": 1}
}' | jq -M .
