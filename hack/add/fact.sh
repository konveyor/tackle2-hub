#!/bin/bash

host="${HOST:-localhost:8080}"

# id (default: 1)
# pass Zero(0) for system assigned.
id="${1:-1}"
key="${2:-pet}"
kind="${3:-dog}"
name="${4:-Rover}"

curl -X PUT ${host}/applications/${id}/facts/${key} \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
-d \
"
kind: ${kind}
name: ${name}
age: 4
"
