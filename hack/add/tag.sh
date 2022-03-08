#!/bin/bash

host="${HOST:-localhost:8080}"

#
# Types
#

curl -X POST ${host}/tagtypes -d \
'{
    "name":"Testing",
    "colour": "#807ded",
    "rank": 0
}' | jq -M .

#
# Tags
#

curl -X POST ${host}/tags -d \
'{
    "username": "tackle",
    "name":"RHEL",
    "tagType": {"id":1}
}' | jq -M .

