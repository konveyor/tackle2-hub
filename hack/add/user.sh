#!/bin/bash

host="${HOST:-localhost:8080}"
userid="${1:-$(whoami)}"
password="${2:-$(whoami)}"
roleId="${3:-3}" # migrator

curl -Ss -k -X POST ${host}/users -H "Authorization: Bearer ${TOKEN}" \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
  -d \
"
userid: ${userid}
password: ${password}
email: $(whoami)@redhat.com
roles:
- id: ${roleId}
"

