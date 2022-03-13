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
        "name": "Test",
        "locator": "grp.app.1.test",
        "data": {
          "application": 1
	}
      }
    ]
}' | jq -M .

