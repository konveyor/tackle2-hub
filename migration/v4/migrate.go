package v4

import (
	"encoding/json"
	"fmt"

	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	v3 "github.com/konveyor/tackle2-hub/migration/v3/model"
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

	err = r.migrateRuleBundles(db)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = r.migrateIdentitiesUniqName(db)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}

//
// RuleBundles renamed: RuleSet
// RuleSet renamed: Rule
func (r Migration) migrateRuleBundles(db *gorm.DB) (err error) {
	err = db.Migrator().RenameTable(&model.RuleSet{}, "RuleSet__old")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.AutoMigrate(&model.RuleSet{}, &model.Rule{})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.Exec("INSERT INTO RuleSet SELECT * FROM RuleBundle").Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	type MD struct {
		Source string `json:"source,omitempty"`
		Target string `json:"target,omitempty"`
	}
	var rulesets []v3.RuleSet
	result := db.Table("RuleSet__old").Find(&rulesets)
	if result.Error != nil {
		err = result.Error
		return
	}
	for _, r := range rulesets {
		labels := []string{}
		md := MD{}
		if r.Metadata != nil {
			err = json.Unmarshal(r.Metadata, &md)
			if err == nil {
				if md.Source != "" {
					labels = append(labels, "konveyor.io/source="+md.Source)
				}
				if md.Target != "" {
					labels = append(labels, "konveyor.io/target="+md.Target)
				}
			}
		}
		rule := &model.Rule{}
		rule.ID = r.ID
		rule.RuleSetID = r.RuleBundleID
		rule.Labels, _ = json.Marshal(labels)
		rule.FileID = r.FileID
		err = db.Create(rule).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	err = db.Exec("UPDATE Setting SET Key = 'ui.ruleset.order' WHERE Key = 'ui.bundle.order'").Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	err = db.Migrator().DropTable("RuleBundle", "RuleSet__old")
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}

func (r Migration) migrateIdentitiesUniqName(db *gorm.DB) (err error) {
	var identities []v3.Identity
	err = db.Find(&identities).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	for _, identity := range identities {
		var dupCount int64
		err = db.Model(identity).Where("name = ?", identity.Name).Where("id != ?", identity.ID).Count(&dupCount).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}

		if dupCount > 0 {
			identity.Name = fmt.Sprintf("%s-ID%d", identity.Name, identity.ID)
			err = db.Save(&identity).Error
			if err != nil {
				err = liberr.Wrap(err)
				return
			}
		}
	}
	return
}
