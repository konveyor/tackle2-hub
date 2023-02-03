#!/bin/bash

host="${HOST:-localhost:8080}"

path="/etc/hosts"

curl -F 'file=@/etc/hosts' http://${host}/files/hosts | jq .

