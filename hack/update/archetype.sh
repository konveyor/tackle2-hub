#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-1}"
kind="${2:-Test}"
name="${3:-Test}-${id}"
genA="${4:-1}"
genB="${5:-2}"
genC="${6:-3}"

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
      - id: ${genA}
      - id: ${genB}
      - id: ${genC}
  - name: other
    generators:
      - id: ${genA}
      - id: ${genB}
"
