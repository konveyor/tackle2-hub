#!/bin/bash

host="${HOST:-localhost:8080}"

#
# Categories
#

curl -X POST ${host}/tagcategories -d \
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
    "category": {"id":1}
}' | jq -M .

