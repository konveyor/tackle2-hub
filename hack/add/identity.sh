#!/bin/bash

host="${HOST:-localhost:8080}"

# id (default: 1)
# pass Zero(0) for system assigned.
id="${1:-1}"
name="${2:-Test}"

# create identity.
curl -X POST ${host}/identities \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
id: ${id}
name: ${name}
kind: source
description: ${name} Description
user: userA
password: passwordA
key: keyA
settings: settingsA
"
