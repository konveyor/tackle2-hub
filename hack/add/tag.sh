#!/bin/bash

host="${HOST:-localhost:8080}"

#
# Types
#

curl -X POST ${host}/controls/tag-type -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"Testing",
    "colour": "#807ded",
    "rank": 0
}' | jq -M .

curl -X POST ${host}/controls/tag-type -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"Operating System",
    "colour": "#807ded",
    "rank": 10
}' | jq -M .

curl -X POST ${host}/controls/tag-type -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"Database",
    "colour": "#8aed7d",
    "rank": 20
}' | jq -M .

curl -X POST ${host}/controls/tag-type -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"Language",
    "colour": "#ede97d",
    "rank": 30
}' | jq -M .

#
# Tags
#

curl -X POST ${host}/controls/tag -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"RHEL",
    "tagType": {"id":1}
}' | jq -M .

curl -X POST ${host}/controls/tag -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"PostgreSQL",
    "tagType": {"id":2}
}' | jq -M .

curl -X POST ${host}/controls/tag -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"C++",
    "tagType": {"id":3}
}' | jq -M .

curl -X POST ${host}/controls/tag -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"CRAZY",
    "tagType": {
      "createUser": "tackle",
      "username": "tackle",
      "name":"CRAZY",
      "colour": "#0000",
      "rank": 40
    }
}' | jq -M .

curl -X POST ${host}/controls/tag -d \
'{
    "createUser": "tackle",
    "username": "tackle",
    "name":"CRAZY-TRAIN",
    "tagType": {
      "id": 4,
      "createUser": "tackle",
      "username": "tackle",
      "name":"CRAZY-TRAIN",
      "colour": "#66666",
      "rank": 40
    }
}' | jq -M .

