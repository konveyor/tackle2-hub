#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=system-assigned.
kind="${2:-Test}"
name="${3:-Test}-${id}"
genid="${4:-1}"

# create archetype.
curl -X POST ${host}/archetypes \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
id: ${id}
name: ${name}
profiles:
  - name: openshift
    generators:
      - id: ${genid}
"
