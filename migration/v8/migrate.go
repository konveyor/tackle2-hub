package v8

import (
	"encoding/json"
	liberr "github.com/jortel/go-utils/error"
	"github.com/jortel/go-utils/logr"
	v7 "github.com/konveyor/tackle2-hub/migration/v7/model"
	"github.com/konveyor/tackle2-hub/migration/v8/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v8")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	result := db.Model(model.Setting{}).Where("key = ?", "ui.ruleset.order").Update("key", "ui.target.order")
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}

	oldCustomRuleSets := []v7.RuleSet{}
	result = db.Find(&oldCustomRuleSets, "uuid IS NULL")
	if result.Error != nil {
		err = liberr.Wrap(err)
		return
	}

	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}

	for _, rs := range oldCustomRuleSets {
		target := model.Target{
			Name:        rs.Name,
			Description: rs.Description,
			RuleSetID:   &rs.ID,
			ImageID:     rs.ImageID,
		}
		target.CreateUser = rs.CreateUser

		type TargetLabel struct {
			Name  string `json:"name"`
			Label string `json:"label"`
		}

		uniqueLabels := make(map[string]bool)
		for _, rule := range rs.Rules {
			ruleLabels := []string{}
			_ = json.Unmarshal(rule.Labels, &ruleLabels)
			for _, label := range ruleLabels {
				uniqueLabels[label] = true
			}
		}

		targetLabels := []TargetLabel{}
		for k, _ := range uniqueLabels {
			targetLabels = append(targetLabels, TargetLabel{Name: k, Label: k})
		}
		target.Labels, _ = json.Marshal(targetLabels)
		result = db.Save(target)
		if result.Error != nil {
			err = liberr.Wrap(result.Error)
			return
		}
	}

	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}
