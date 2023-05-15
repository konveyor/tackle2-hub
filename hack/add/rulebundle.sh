#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/rulebundles -d \
'{
    "name": "bundle-A",
    "description":"Test",
    "image": {"id":1},
    "rulesets": [
      {"name":"RuleSet-1", "metadata": {"target":"Tar-1"}, "file":{"id":2}},
      {"name":"RuleSet-2", "metadata": {"target":"Tar-2"}, "file":{"id":3}},
      {"name":"RuleSet-3", "Rules": [
           {"name":"rule-1","file":{"id":2}},
           {"name":"rule-2","file":{"id":3}}
         ]}
    ]
}' | jq .

