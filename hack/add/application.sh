#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=system-assigned.
name="${2:-Test}"
repository="${2:-https://github.com/WASdev/sample.daytrader7.git}"
businessService="${3:-1}"

# create application.
curl -X POST ${host}/applications \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
id: ${id}
name: ${name}
description: ${name} application.
businessService:
  id: ${businessService}
repository:
    kind: git
    url: ${repository}
identities:
  - id: 1
tags:
  - id: 1
  - id: 16
"
