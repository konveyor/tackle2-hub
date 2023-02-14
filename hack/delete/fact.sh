#!/bin/bash

host="${HOST:-localhost:8080}"

curl -X DELETE ${host}/applications/1/facts/address
