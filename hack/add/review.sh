#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=assigned
application="${2:-1}"

curl -X POST ${host}/reviews \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
-d \
"
id: ${id}
businessCriticality: 4
effortEstimate: large
proposedAction: proceed
workPriority: 1
comments: This is good.
application:
  id: ${application}
"
