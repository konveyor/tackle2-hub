#!/bin/bash

host="${HOST:-localhost:8080}"

# id (default: 1)
# pass Zero(0) for system assigned.
id="${1:-1}"
from="${2:-2}"
to="${3:-1}"

curl -X POST ${host}/dependencies \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
  -d \
"
from:
  id: ${from}
to:
  id: ${to}
"
