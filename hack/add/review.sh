#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/application-inventory/review -d \
'{
    "createUser": "tackle",
    "businessCriticality": 4,
    "effortEstimate": "high",
    "proposedAction": "proceed",
    "workPriority": 1,
    "comments": "This is good.",
    "application": {
      "id": 1
    }
}' | jq -M .
