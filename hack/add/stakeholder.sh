#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/stakeholders -d \
'{
    "name": "tackle",
    "email": "tackle@konveyor.org",
    "stakeholderGroups": [{"id": 1}],
    "jobFunction" : {"id": 1}
}' | jq -M .
