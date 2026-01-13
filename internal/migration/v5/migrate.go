package v5

import (
	"encoding/json"
	"fmt"

	liberr "github.com/jortel/go-utils/error"
	v3 "github.com/konveyor/tackle2-hub/internal/migration/v3/model"
	v4 "github.com/konveyor/tackle2-hub/internal/migration/v4/model"
	"github.com/konveyor/tackle2-hub/internal/migration/v5/model"
	"gorm.io/gorm"
)

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
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
	err = r.updateBundleSeed(db)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	return
}

func (r Migration) Models() []any {
	return model.All()
}

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
	var rulesets []v4.RuleSet
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

// updateBundleSeed updates the description for Open Liberty.
func (r Migration) updateBundleSeed(db *gorm.DB) (err error) {
	db = db.Model(&model.RuleSet{})
	db = db.Where("Name", "Open Liberty")
	err = db.Update(
		"Description",
		"A comprehensive set of rules for migrating traditional WebSphere"+
			" applications to Open Liberty.").Error
	return
}

// migrateIdentitiesUniqName de-duplicates identity names.
func (r Migration) migrateIdentitiesUniqName(db *gorm.DB) (err error) {
	var identities []v3.Identity
	err = db.Find(&identities).Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	dupes := make(map[string]int)
	for i := range identities {
		id := &identities[i]
		dupes[id.Name]++
	}

	for _, id := range identities {
		if dupes[id.Name] < 2 {
			continue
		}

		suffix := 0
		for {
			suffix++
			newName := fmt.Sprintf("%s (%d)", id.Name, suffix)
			_, found := dupes[newName]
			if !found {
				id.Name = newName
				dupes[newName] = 1
				break
			}
		}
		err = db.Save(id).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	return
}
