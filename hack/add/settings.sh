#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/settings -d '{"key":"test.boolean","value":false}' | jq .

