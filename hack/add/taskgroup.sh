#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X POST ${host}/taskgroups -d \
'{
    "name": "Test",
    "addon": "test",
    "data": {
      "path": "/etc"
    },
    "tasks": [
      {
        "name": "Test-1",
	"application": {"id": 1},
        "data": {
	}
      },
      {
        "name": "Test-2",
        "application": {"id": 1},
        "data": {
        }
      }
    ]
}' | jq -M .

