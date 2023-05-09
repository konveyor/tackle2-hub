#!/bin/bash

host="localhost:8080"

#######################################################
# ALL
#######################################################

dir=`dirname $0`
cd ${dir}

./settings.sh
./tag.sh
./identity.sh
./job-function.sh
./stakeholder-group.sh
./stakeholder.sh
./business-service.sh
./application.sh
./task.sh
./taskgroup.sh
./bucket.sh
./review.sh
./proxy.sh
./applicationtags.sh
./analysis.sh
./analysis.sh
./analysis.sh
