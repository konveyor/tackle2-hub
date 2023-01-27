#!/bin/bash

host="${HOST:-localhost:8080}"
user="${1:-admin}"
password="${2:-admin}"

curl -i -X POST ${host}/auth/login -d "{\"user\":\"${user}\",\"password\":\"${password}\"}"

