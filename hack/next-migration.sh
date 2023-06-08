#!/bin/bash

set -e

root="migration"

#
# Determine migration versions.
#
migrations=($(find ${root} -maxdepth 1 -type d -name  'v*' -printf '%f\n' | sort))
current=${migrations[-1]}
n=${current#"v"}

current="v${n}"
((n++))
next="v${n}"

currentDir="${root}/${current}"
nextDir="${root}/${next}"

echo "Current: ${currentDir}"
echo "Next:    ${nextDir}"

#
# Create directores.
#
mkdir -p ${nextDir}/model

#
# Build migrate.go
#
file=${nextDir}/migrate.go
migrate=$(cat << EOF
package ${next}

import (
        "github.com/jortel/go-utils/logr"
//      "github.com/konveyor/tackle2-hub/migration/${next}/model"
        "gorm.io/gorm"
)

var log = logr.WithName("migration|${current}")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
        return
}

EOF
)

echo "${migrate}" > ${file}

#
# Build model/pkg.go
#
file=${nextDir}/model/pkg.go
pkg=$(cat << EOF
package model

import "github.com/konveyor/tackle2-hub/migration/${current}/model"
EOF
)

echo "${pkg}" > ${file}
echo "" >> ${file}

grep "type" model/pkg.go | grep "model" | sort  >> ${file}

#
# Point model at new migration.
#
sed -i "s/${current}/${next}/g" model/pkg.go

#
# DONE
#
echo "Done!"

