#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/application-inventory/application -d \
'{
    "createUser": "tackle",
    "name":"Dog",
    "description": "Dog application.",
    "businessService": "1",
    "tags":[
      "1",
      "2"
    ]
}' | jq -M .

curl -X POST ${host}/application-inventory/application -d \
'{
    "createUser": "tackle",
    "name":"Cat",
    "description": "Cat application.",
    "repository": {
      "name": "Cat",
      "kind": "git",
      "url": "git://github.com/pet/cat",
      "branch": "/cat"
    },
    "businessService": "1",
    "tags":[
      "1",
      "2"
    ]
}' | jq -M .

curl -X POST ${host}/application-inventory/application -d \
'{
    "createUser": "tackle",
    "name":"Pathfinder",
    "description": "Tackle Pathfinder application.",
    "repository": {
      "name": "tackle-pathfinder",
      "url": "https://github.com/konveyor/tackle2-pathfinder.git",
      "branch": "1.2.0"
    },
    "businessService": "1"
}' | jq -M .

