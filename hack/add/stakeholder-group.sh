#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}"
# pass Zero(0) for system assigned.
id="${1:-0}"
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

