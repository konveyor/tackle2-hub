#!/bin/bash

host="${HOST:-localhost:8080}"

#
# Tags
#

curl -X PUT ${host}/tags/2 -d \
'{
   "username": "tackle",
   "name":"Windows",
   "tagType": {"id":1}
}' | jq -M .
