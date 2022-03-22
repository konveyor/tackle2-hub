#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/dependencies -d \
'{
    "from": {"id": 1},
    "to": {"id": 3}
}' | jq -M .
