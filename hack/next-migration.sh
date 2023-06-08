#!/bin/bash

set -e

root="migration"
importRoot="github.com/konveyor/tackle2-hub/migration"

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
//      "${importRoot}/${next}/model"
        "gorm.io/gorm"
)

var log = logr.WithName("migration|${current}")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
        return
}

func (r Migration) Models() []interface{} {
	return model.All()
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

import "${importRoot}/${current}/model"
EOF
)

echo "${pkg}" > ${file}
echo "" >> ${file}

grep "type" model/pkg.go | grep "model" | sort  >> ${file}

#
# Register new migration.
#

sed -i "s|${current} \"${importRoot}/${current}\"|${current} \"${importRoot}/${current}\"\n\t${next} \"${importRoot}/${next}\"|g" ${root}/pkg.go
sed -i "s|${current}.Migration{}|${current}.Migration{},\n\t\t${next}.Migration{}|g" ${root}/pkg.go

#
# Point model at new migration.
#
sed -i "s/${current}/${next}/g" model/pkg.go

