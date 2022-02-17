#!/bin/bash

host="localhost:8080"

#######################################################
# ALL
#######################################################

dir=`dirname $0`
cd ${dir}

./tag.sh
./job-function.sh
./stakeholder-group.sh
./stakeholder.sh
./business-service.sh
./application.sh
./review.sh
./identity.sh
./proxy.sh

