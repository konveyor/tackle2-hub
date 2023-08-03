#!/bin/bash

host="${HOST:-localhost:8080}"
file="${1:-1}"

curl -X POST ${host}/rulesets \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
---
name: Test
description: Test ruleset.
rules:
 - name: Example
   file:
     id: ${file}
"
