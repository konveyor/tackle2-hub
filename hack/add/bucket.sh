#!/bin/bash

host="${HOST:-localhost:8080}"

path="/etc/hosts"

echo "______________________ BUCKET ______________________"
curl  -F 'file=@/etc/hosts' ${host}/buckets/1/${path}
curl  -X GET ${host}/buckets/1/${path}
curl  -X DELETE ${host}/buckets/1/${path}
curl  -X GET ${host}/buckets/1/${path}

echo "______________________ APP ______________________"
curl  -F 'file=@/etc/hosts' ${host}/applications/1/bucket/${path}
curl  -X GET ${host}/applications/1/bucket/${path}
curl  -X DELETE ${host}/applications/1/bucket/${path}
curl  -X GET ${host}/applications/1/bucket/${path}

echo "______________________ TASK ______________________"
curl  -F 'file=@/etc/hosts' ${host}/tasks/1/bucket/${path}
curl  -X GET ${host}/tasks/1/bucket/${path}
curl  -X DELETE ${host}/tasks/1/bucket/${path}
curl  -X GET ${host}/tasks/1/bucket/${path}

echo "______________________ TASKGROUP ______________________"
curl  -F 'file=@/etc/hosts' ${host}/taskgroups/1/bucket/${path}
curl  -X GET ${host}/taskgroups/1/bucket/${path}
curl  -X DELETE ${host}/taskgroups/1/bucket/${path}
curl  -X GET ${host}/taskgroups/1/bucket/${path}
