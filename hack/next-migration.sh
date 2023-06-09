#!/bin/bash
#
# - Create migration/vX
# - Create migration/vX/model
# - Build migration/vX/model/pkg.go
# - Build migration/vX/migrate.go
# - Edit migration/migrate.go add vX to migrations array
# - Edit model/plg.go to import from migration vX
#

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
	"${importRoot}/${next}/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|${current}")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	err = db.AutoMigrate(r.Models()...)
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
models=$(grep "type" model/pkg.go | grep "model")
echo "${models}"  >> ${file}
echo -n "
//
// All builds all models.
// Models are enumerated such that each are listed after
// all the other models on which they may depend.
func All() []interface{} {
	return []interface{}{
" >> ${file}
echo "${models}" | grep -v "Model" | while read m
do
  part=(${m})
  m=${part[3]#"model."}
  echo -e "\t\t${m}{}," >> ${file}
done
echo -e "\t}" >> ${file}
echo "}" >> ${file}

#
# Register new migration.
#

sed -i "s|${current} \"${importRoot}/${current}\"|${current} \"${importRoot}/${current}\"\n\t${next} \"${importRoot}/${next}\"|g" ${root}/pkg.go
sed -i "s|${current}.Migration{}|${current}.Migration{},\n\t\t${next}.Migration{}|g" ${root}/pkg.go

#
# Point model at new migration.
#
sed -i "s/${current}/${next}/g" model/pkg.go

