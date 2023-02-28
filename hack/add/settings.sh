#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/settings -d '{"key":"test.boolean","value":false}' | jq .
curl -X POST ${host}/settings -d '{"key":"test.numbers","value":[1,2,3]}' | jq .
curl -X POST ${host}/settings/test.bykey -d '123'