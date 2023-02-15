#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT "${host}/applications/1/tags?source=analysis" -d \
'[{"id":1},{"id":2},{"id":3},{"id":4}]' | jq -M .

curl -X POST "${host}/applications/1/tags" -d '{"id":4}' | jq -M .