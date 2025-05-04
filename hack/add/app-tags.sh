#!/bin/bash

host="${HOST:-localhost:8080}"

# Replace analysis tags.
curl -X PUT "${host}/applications/1/tags?source=analysis" \
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
curl -X POST "${host}/applications/1/tags" \
   -H 'Content-Type:application/x-yaml' \
   -H 'Accept:application/x-yaml' \
  -d \
 '
 id: 4
 '