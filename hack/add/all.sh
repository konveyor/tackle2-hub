#!/bin/bash

dir=`dirname $0`
cd ${dir}

./settings.sh
./identity.sh
./job-function.sh
./stakeholder-group.sh
./stakeholder.sh
./business-service.sh
./application.sh
./dependency.sh
./fact.sh
./review.sh
./tag-category.sh
./tag.sh
./file.sh
./ruleset.sh
