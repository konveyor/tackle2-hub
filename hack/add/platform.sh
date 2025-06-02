#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=system-assigned.
kind="${2:-Test}"
name="${3:-Test}"

# create application.
curl -X POST ${host}/platforms \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
id: ${id}
kind: ${kind}
name: ${name}
url: http://platform.org
"
