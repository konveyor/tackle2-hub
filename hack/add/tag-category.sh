#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=system-assigned.
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


