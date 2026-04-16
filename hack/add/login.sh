#!/bin/bash

host="${HOST:-localhost:8080}"
user="${1:-admin}"
password="${2:-admin}"
lifespan="${3:-24}"

curl -Ss -k -X POST ${host}/auth/tokens \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
  -d \
"
userid: ${user}
password: ${password}
lifespan: ${lifespan}
"

