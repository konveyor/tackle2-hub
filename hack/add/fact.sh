#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/applications/1/facts/pet -d \ '{"kind":"dog","Age":4}' | jq -M .
curl -X PUT ${host}/applications/1/facts/address -d \ '{"street":"Maple","State":"AL"}' | jq -M .
