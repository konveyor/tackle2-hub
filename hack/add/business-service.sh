#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}"
# pass Zero(0) for system assigned.
id="${1:-0}"
name="${2:-Test}"
owner="${3:-1}"

# Create a business service.
curl -X POST ${host}/businessservices \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
-d \
"
id: ${id}
name: ${name}
description: ${name} Dept.
owner:
  id: ${owner}
"