#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/applications/1/facts/address \
  -H 'Content-Type:application/x-yaml' \
  -H 'Accept:application/x-yaml' \
  -d \
'
---
value:
  street: 1234 Maple St.
  city: Huntsville
  state: AL
  zip: 35763
'
