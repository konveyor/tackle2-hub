#!/bin/bash
#
# Usage: jwt.sh <key>
#
key=$1
header='{"alg":"HS512","typ":"JWT"}'
payload='{"scope":"applicaions:get","user":"operator"}'
headerStr=$(echo -n ${header} \
  | base64 -w 0 \
  | sed s/\+/-/g \
  | sed 's/\//_/g' \
  | sed -E s/=+$//)
payloadStr=$(echo -n ${payload} \
  | base64 -w 0 \
  | sed s/\+/-/g \
  | sed 's/\//_/g' \
  | sed -E s/=+$//)
signStr=$(echo -n "${headerStr}.${payloadStr}" \
  | openssl dgst -sha512 -mac HMAC -macopt key:${key} -binary \
  | base64  -w 0 \
  | sed s/\+/-/g \
  | sed 's/\//_/g' \
  | sed -E s/=+$//)
token="${headerStr}.${payloadStr}.${signStr}"
echo "${token}"
