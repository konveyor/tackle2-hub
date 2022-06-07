#!/bin/bash

host="${HOST:-localhost:8080}"

# ID to update (default:1)
id="${1:-1}"


curl -I -X PUT ${host}/tasks/${id}/cancel
