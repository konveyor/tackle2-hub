#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT "${host}/applications/1/tags?source=analysis" -d \
'[{"id":5},{"id":6},{"id":7},{"id":8}]' | jq -M .
