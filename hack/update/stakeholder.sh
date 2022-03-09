#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/stakeholders/1 -d \
'{
    "id": 99,
    "name": "Fudd",
    "email": "tackle@konveyor.org",
    "stakeholderGroups": [{"id": 1}],
    "jobFunction" : {"id": 1}
}' | jq -M .
