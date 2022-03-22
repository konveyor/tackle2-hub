#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X PUT ${host}/taskgroups/1/submit -d \
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
        "name": "Renamed",
        "locator": "renamed",
        "data": {
          "application": 1
	}
      },
      {
        "name": "Renamed",
        "locator": "renamed2",
        "data": {
	  "x": 1
        }
      }
    ]
}' | jq -M .

