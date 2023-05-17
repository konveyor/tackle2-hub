package v4

import (
	"encoding/json"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v4/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v4")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	// Adding Source column to Fact composite primary key.
	// Altering the primary key requires constructing a new table, so rename the old one,
	// create the new one, copy over the rows, and then drop the old one.
	err = db.Migrator().RenameTable("Fact", "Fact__old")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.Migrator().CreateTable(model.Fact{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	result := db.Exec("INSERT INTO Fact (ApplicationID, Key, Value, Source) SELECT ApplicationID, Key, Value, '' FROM Fact__old;")
	if result.Error != nil {
		err = liberr.Wrap(result.Error)
		return
	}
	err = db.Migrator().DropTable("Fact__old")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = r.addLabels(db)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}

func (r Migration) addLabels(db *gorm.DB) (err error) {
	type MD struct {
		Source string `json:"source,omitempty"`
		Target string `json:"target,omitempty"`
	}
	var rulesets []model.RuleSet
	result := db.Find(&rulesets)
	if result.Error != nil {
		err = result.Error
		return
	}
	for _, r := range rulesets {
		var labels []string
		md := MD{}
		if r.Metadata == nil {
			continue
		}
		err = json.Unmarshal(r.Metadata, &md)
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
		if md.Source != "" {
			labels = append(labels, "konveyor.io/source="+md.Source)
		}
		if md.Target != "" {
			labels = append(labels, "konveyor.io/target="+md.Target)
		}
		if len(labels) > 0 {
			r.Labels, _ = json.Marshal(labels)
			err = db.Save(r).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	return
}
