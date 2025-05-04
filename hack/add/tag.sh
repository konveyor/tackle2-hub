#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=assigned
name="${2:-Test}"
category="${3:-1}"

# create category.
curl -X POST ${host}/tags \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
-d \
"
id: ${id}
name: ${name}
category:
  id: ${category}
"