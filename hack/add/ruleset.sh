#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=assigned
file="${2:-1}"

curl -X POST ${host}/rulesets \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
---
id: ${id}
name: Test
description: Test ruleset.
rules:
 - name: Example
   file:
     id: ${file}
"
