#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/applications -d \
'{
    "name":"Dog",
    "description": "Dog application.",
    "businessService": {"id":1},
    "identities": [
      {"id":1},
      {"id":2}
    ],
    "tags":[
      {"id":1}
    ]
}' | jq -M .

curl -X POST ${host}/applications -d \
'{
    "name":"Cat",
    "description": "Cat application.",
    "repository": {
      "name": "Cat",
      "kind": "git",
      "url": "git://github.com/pet/cat",
      "branch": "/cat"
    },
    "businessService": {"id":1},
    "tags":[
      {"id":1}
    ]
}' | jq -M .

curl -X POST ${host}/applications -d \
'{
    "createUser": "tackle",
    "name":"Pathfinder",
    "description": "Tackle Pathfinder application.",
    "repository": {
      "name": "tackle-pathfinder",
      "url": "https://github.com/konveyor/tackle2-pathfinder.git",
      "branch": "1.2.0"
    },
    "facts": {
      "analysed": true
    },
    "businessService": {"id":1}
}' | jq -M .

