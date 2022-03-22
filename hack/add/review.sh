#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/reviews -d \
'{
    "businessCriticality": 4,
    "effortEstimate": "large",
    "proposedAction": "proceed",
    "workPriority": 1,
    "comments": "This is good.",
    "application": {"id":1}
}' | jq -M .

curl -X POST ${host}/reviews -d \
'{
    "businessCriticality": 4,
    "effortEstimate": "small",
    "proposedAction": "rehost",
    "workPriority": 1,
    "comments": "This is different.",
    "application": {"id":2}
}' | jq -M .

curl -X POST ${host}/reviews -d \
'{
    "businessCriticality": 4,
    "effortEstimate": "extra_large",
    "proposedAction": "repurchase",
    "workPriority": 1,
    "comments": "This is hard.",
    "application": {"id":3}
}' | jq -M .
