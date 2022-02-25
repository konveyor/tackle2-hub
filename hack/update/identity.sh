#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/identities/1 -d \
'{
    "kind": "git",
    "name":"test-git",
    "description": "Forklift",
    "user": "userB",
    "password": "passwordB",
    "key": "keyA",
    "settings": "settingsB"
}' | jq -M .

