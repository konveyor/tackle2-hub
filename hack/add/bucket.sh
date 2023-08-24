#!/bin/bash

host="${HOST:-localhost:8080}"
id="${1:-1}"
path="${2:-/etc/hosts}"

echo "______________________ BUCKET ______________________"
curl  -F "file=@${path}" ${host}/buckets/${id}/${path}
curl  --output /tmp/get-bucket -X GET ${host}/buckets/${id}/${path}
curl  -X DELETE ${host}/buckets/${id}/${path}
curl  -X GET ${host}/buckets/${id}/${path}

echo "______________________ APP ______________________"
curl  -F "file=@${path}" ${host}/applications/${id}/bucket/${path}
curl  --output /tmp/get-bucket-app -X GET ${host}/applications/${id}/bucket/${path}
curl  -X DELETE ${host}/applications/${id}/bucket/${path}
curl  -X GET ${host}/applications/${id}/bucket/${path}

echo "______________________ TASK ______________________"
curl  -F "file=@${path}" ${host}/tasks/${id}/bucket/${path}
curl  --output /tmp/get-bucket-task -X GET ${host}/tasks/${id}/bucket/${path}
curl  -X DELETE ${host}/tasks/${id}/bucket/${path}
curl  -X GET ${host}/tasks/${id}/bucket/${path}

echo "______________________ TASKGROUP ______________________"
curl  -F "file=@${path}" ${host}/taskgroups/${id}/bucket/${path}
curl  --output /tmp/get-bucket-tg -X GET ${host}/taskgroups/${id}/bucket/${path}
curl  -X DELETE ${host}/taskgroups/${id}/bucket/${path}
curl  -X GET ${host}/taskgroups/${id}/bucket/${path}
