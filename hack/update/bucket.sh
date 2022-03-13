#!/bin/bash

set -x
set -e

host="${HOST:-localhost:8080}"

# ID to update (default:1)
id="${1:-1}"

path="${2:-/etc/hosts}"

# application
curl -F "file=@${path}" ${host}/applications/${id}/bucket${path}
curl ${host}/applications/${id}/bucket${path}

# task group.
curl -F "file=@${path}" ${host}/taskgroups/${id}/bucket${path}
curl ${host}/applications/${id}/bucket${path}

# task.
curl -F "file=@${path}" ${host}/tasks/${id}/bucket${path}
curl ${host}/tasks/${id}/bucket${path}

