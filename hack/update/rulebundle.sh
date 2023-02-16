#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/rulebundles/12 -d \
'{
    "name": "bundle-A (updatted)",
    "description": "Test (updated)",
    "image": {"id":2},
    "rulesets": [
      {"name":"RuleSet-2", "metadata": {"target":"Tar-2"}, "file":{"id":4}},
      {"name":"RuleSet-3", "metadata": {"target":"Tar-3"}, "file":{"id":5}}
    ]
}' | jq .

