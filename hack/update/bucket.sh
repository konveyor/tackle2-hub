#!/bin/bash

set -x
set -e

host="${HOST:-localhost:8080}"

# ID to update (default:1)
id="${1:-1}"

path="${2:-/etc/hosts}"

archive="${3:-bucket-dir-sample.tar.gz}"

for resource in applications tasks taskgroups; do
    echo
    echo "Updating bucket for ${resource}."
    echo

    # Remove previous working files
    rm -f new-dir.tar.gz
    rm -f bucket.tar.gz

    # Upload and get a single file
    curl -F "file=@${path}" ${host}/${resource}/${id}/bucket${path}
    curl ${host}/${resource}/${id}/bucket${path}   # should display content of the $path file (/etc/host by default)

    # Upload archive to a subdirectory
    curl -v -X PUT -H "X-Directory:expand" -F "file=@${archive}" ${host}/${resource}/${id}/bucket/new-dir

    # Download new-dir as an archive
    curl -vOJ -H "X-Directory:archive" ${host}/${resource}/${id}/bucket/new-dir    # should create local file new-dir.tar.gz

    # Display index page for the bucket
    curl ${host}/${resource}/${id}/bucket/

    # And download the bucket as an archive
    curl -v -H "X-Directory:archive" -o bucket.tar.gz ${host}/${resource}/${id}/bucket/    # should create local file bucket.tar.gz (<bucket-uuid>.tar.gz when no destination was set)
done
