#!/bin/bash

host="${HOST:-localhost:8080}"
appid="${1:-1}"

# Replace analysis tags.
curl -X PUT "${host}/applications/${appid}/tags?source=analysis" \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
'
- id: 1
- id: 2
- id: 3
- id: 4
'

# add user tag.
curl -X POST "${host}/applications/${appid}/tags" \
   -H 'Content-Type:application/x-yaml' \
   -H 'Accept:application/x-yaml' \
  -d \
 '
 id: 4
 '