#!/bin/bash

host="${OIDC:-localhost:8080}"
user="$1"
password="$2"

curl -k -X POST ${host}/realms/tackle/protocol/openid-connect/token \
	--user backend-service:secret \
	-H "content-type: application/x-www-form-urlencoded" \
	-d "username=${user}&password=${password}&grant_type=password&client_id=tackle-ui" \
	| jq --raw-output ".access_token"
