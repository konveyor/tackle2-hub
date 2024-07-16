#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/identities -d \
'{
    "kind": "source",
    "name":"test-git",
    "description": "Forklift",
    "user": "userA",
    "password": "passwordA",
    "key": "keyA",
    "settings": "settingsA"
}' | jq -M .

curl -X POST ${host}/identities -d \
'{
    "kind": "maven",
    "name":"test-mvn",
    "description": "Forklift",
    "user": "userA",
    "password": "passwordA",
    "key": "keyA",
    "settings": "settingsA"
}' | jq -M .
