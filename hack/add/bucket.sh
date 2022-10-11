#!/bin/bash

host="${HOST:-localhost:8080}"

path="/etc/hosts"

echo "______________________ APP ______________________"
curl -F 'file=@/etc/hosts' http://${host}/applications/1/bucket/${path}
curl -X GET http://${host}/applications/1/bucket/${path}
curl -X DELETE http://${host}/applications/1/bucket/${path}
curl -X GET http://${host}/applications/1/bucket/${path}

echo "______________________ TASK ______________________"
curl -F 'file=@/etc/hosts' http://${host}/tasks/1/bucket/${path}
curl -X GET http://${host}/tasks/1/bucket/${path}
curl -X DELETE http://${host}/tasks/1/bucket/${path}
curl -X GET http://${host}/tasks/1/bucket/${path}

echo "______________________ TASKGROUP ______________________"
curl -F 'file=@/etc/hosts' http://${host}/taskgroups/1/bucket/${path}
curl -X GET http://${host}/taskgroups/1/bucket/${path}
curl -X DELETE http://${host}/taskgroups/1/bucket/${path}
curl -X GET http://${host}/taskgroups/1/bucket/${path}
