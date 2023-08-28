#!/bin/bash

host="localhost:8080"

#######################################################
# ALL
#######################################################

dir=`dirname $0`
cd ${dir}

./settings.sh
./identity.sh
./job-function.sh
./stakeholder-group.sh
./stakeholder.sh
./business-service.sh
./application.sh
./review.sh
./taskgroup.sh
