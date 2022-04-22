#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/taskgroups -d \
'{
    "name": "Test",
    "state": "Ready",
    "addon": "test",
    "data": {
      "path": "/etc"
    },
    "tasks": [
      {
        "name": "Test",
	"application": {"id": 1},
        "data": {
	}
      }
    ]
}' | jq -M .

