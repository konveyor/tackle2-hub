#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/taskgroups/1 -d \
'{
    "id": 1,
    "name": "Test-updated",
    "addon": "test",
    "data": {
      "path": "/etc/updated",
      "application": 2,
      "other": 18
    },
    "tasks": [
      {
        "id": 3,
	"state": "Created",
        "name": "Renamed",
        "locator": "renamed",
        "data": {
          "application": 1
	}
      },
      {
        "name": "Another",
        "locator": "another.2",
        "data": {
          "application": 2
        }
      }
    ]
}' | jq -M .

