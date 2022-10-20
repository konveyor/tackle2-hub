#!/bin/bash

set -x
set -e

host="${HOST:-localhost:8080}"

# ID to update (default:1)
id="${1:-1}"

path="${2:-/etc/hosts}"

archive="bucket-dir-sample.tar.gz"

# Application

# Remove previous working files
rm -f new-dir.tar.gz
rm -f bucket.tar.gz

# Upload and get a single file
curl -F "file=@${path}" ${host}/applications/${id}/bucket${path}
curl ${host}/applications/${id}/bucket${path}   # should display content of the $path file (/etc/host by default)

# Upload archive to a subdirectory
curl -v -X PUT -H "X-Directory:expand" -F "file=@${archive}" ${host}/applications/${id}/bucket/new-dir

# Download new-dir as an archive
curl -vOJ -H "X-Directory:archive" ${host}/applications/${id}/bucket/new-dir    # should create local file new-dir.tar.gz

# Display index page for the bucket
curl ${host}/applications/${id}/bucket/

# And download the bucket as an archive
curl -v -H "X-Directory:archive" -o bucket.tar.gz ${host}/applications/${id}/bucket/    # should create local file bucket.tar.gz (<bucket-uuid>.tar.gz when no destination was set)

