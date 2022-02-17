#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/identities -d \
'{
    "createUser": "tackle",
    "kind": "git",
    "name":"jeff",
    "description": "Forklift",
    "user": "userA",
    "password": "passwordA",
    "key": "keyA",
    "settings": "settingsA",
    "application": 1
}' | jq -M .

curl -X POST ${host}/application-inventory/application/1/identities -d \
'{
    "createUser": "tackle",
    "kind": "mvn",
    "name":"jeff-mvn",
    "description": "Forklift",
    "user": "userA",
    "password": "passwordA",
    "key": "keyA",
    "settings": "settingsA"
}' | jq -M .
