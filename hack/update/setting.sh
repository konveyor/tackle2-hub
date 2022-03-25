#!/bin/bash

host="${HOST:-localhost:8080}"
key=$1
value=$2

curl -X PUT ${host}/settings/${key} -d "${value}"
