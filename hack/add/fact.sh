#!/bin/bash

host="${HOST:-localhost:8080}"

application="${1:-1}"
key="${2:-pet}"
kind="${3:-dog}"
name="${4:-Rover}"

curl -X PUT ${host}/applications/${application}/facts/${key} \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
-d \
"
kind: ${kind}
name: ${name}
age: 4
"
