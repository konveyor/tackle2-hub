#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/identities \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
'
kind: source
name: test-git
description: Test Description
user: userA
password: passwordA
key: keyA
settings: settingsA
'
