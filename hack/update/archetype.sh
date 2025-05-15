#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-1}"
kind="${2:-Test}"
name="${3:-Test}-${id}"
genid="${4:-1}"

# update archetype
curl -X PUT ${host}/archetypes/${id} \
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
  - name: other
    generators:
      - id: ${genid}
"
