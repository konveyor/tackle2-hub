#!/bin/bash

host="${HOST:-localhost:8080}"

id="${1:-0}" # 0=system-assigned.
name="${2:-Test}"
kind="${3:-source}"
def="${4:-0}"

# create identity.
curl -X PUT ${host}/identities/${id} \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
 -d \
"
id: ${id}
name: ${name}
kind: ${kind}
default: ${def}
description: ${name} Description
user: userA
password: passwordA
key: keyA
settings: settingsA
"
