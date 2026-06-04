#!/bin/bash

host="${HOST:-localhost:8080}"
lifespan="${1:-24}"

curl -Ss -k -X POST ${host}/auth/tokens \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
  -H "Authorization:Bearer ${TOKEN}" \
  -d \
"
lifespan: ${lifespan}
"

