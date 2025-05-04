#!/bin/bash

host="${HOST:-localhost:8080}"

# id (default: 0)
# pass Zero(0) for system assigned.
id="${1:-0}"
name="${2:-Test}"

# create category.
curl -X POST ${host}/tagcategories \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
-d \
"
id: ${id}
name: ${name}
colour: #807ded
"


