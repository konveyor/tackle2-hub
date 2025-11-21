#!/bin/bash

host="${HOST:-localhost:8080}"
user="${1:-admin}"
password="${2:-admin}"

curl -Ss -k -X POST ${host}/auth/login \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
  -d \
"
user: ${user}
password: ${password}
"

