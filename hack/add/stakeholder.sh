#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=system-assigned.
name="${2:-Test}"
group="${3:-1}"
jobFunction="${4:-1}"

# create stakeholder.
#
curl -X POST ${host}/stakeholders \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
 "
 id: ${id}
 name: ${name}
 email: tackle@konveyor.org
 stakeholderGroups:
   - id: ${group}
 jobFunction:
   id: ${jobFunction}
 "

