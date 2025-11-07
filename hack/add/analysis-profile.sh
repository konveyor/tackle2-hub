#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=system-assigned.
name="${1:-Test}-${id}"
repository="${2:-https://github.com/WASdev/sample.daytrader7.git}"

# create application.
curl -X POST ${host}/analysis/profiles \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
id: ${id}
name: ${name}
mode:
  withDeps: true
scope:
  withKnownLibs: true
  packages:
    included:
      - one
      - two
    excluded:
      - three
      - four
rules:
  labels:
    included:
      - A
      - B
    excluded:
      - C
      - D
  targets:
    - id: 1
    - id: 2
    - id: 3
  files:
    - id: 400
  repository:
    kind: git
    url: ${repository}
"
