#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0 = system assigned.
name="${2:-Test}"

#
# Create a stakeholder group.
#
curl -X POST ${host}/stakeholdergroups \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
 "
 id: ${id}
 name: ${name}
 description: ${name} group.
 "

