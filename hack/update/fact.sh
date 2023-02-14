#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/applications/1/facts/address -d \ '{"street": "Maple","City":"Huntsville","State":"AL"}' | jq -M .
