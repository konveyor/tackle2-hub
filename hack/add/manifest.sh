#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=system-assigned.
appid="${2:-1}"

curl -X POST ${host}/manifests \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
---
id: ${id}
application:
  id: ${appid}
content:
  name: Test
  service:
    name: test-service
    other: 10
  deployment:
    name: test-deployment
    other: 20
  password: \$(password)
secret:
  user: Elmer
  password: rabbit-slayer
"
