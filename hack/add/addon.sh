#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/addons/test/tasks -d \
'{
   "application": 1,
   "path": "/etc",
   "delay": 5
}' | jq -M . 

