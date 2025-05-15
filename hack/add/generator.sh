#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=system-assigned.
kind="${2:-Test}"
name="${3:-Test}-${id}"
repository="${4:-https://github.com/WASdev/sample.daytrader7.git}"

# create application.
curl -X POST ${host}/generators \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
id: ${id}
kind: ${kind}
name: ${name}
repository:
    kind: git
    url: ${repository}
"
