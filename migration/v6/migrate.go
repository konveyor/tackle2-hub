package v6

import (
	"encoding/json"

	"github.com/jortel/go-utils/logr"
	"github.com/konveyor/tackle2-hub/migration/v6/model"
	"gorm.io/gorm"
)

var log = logr.WithName("migration|v6")

type Migration struct{}

func (r Migration) Apply(db *gorm.DB) (err error) {
	m := db.Migrator()
	err = m.DropIndex(model.TechDependency{}, "depA")
	if err != nil {
		return
	}
	err = m.DropIndex(model.Issue{}, "issueA")
	if err != nil {
		return
	}
	err = db.AutoMigrate(r.Models()...)
	if err != nil {
		return
	}
	err = r.taskReportError(db)
	if err != nil {
		return
	}
	err = r.taskError(db)
	if err != nil {
		return
	}
	return
}

func (r Migration) Models() []interface{} {
	return model.All()
}

func (r Migration) taskError(db *gorm.DB) (err error) {
	type Task struct {
		model.Task
		Error string
	}
	var list []Task
	err = db.Find(&Task{}, &list).Error
	if err != nil {
		return
	}
	for i := range list {
		m := &list[i]
		if m.Error == "" {
			continue
		}
		m.Errors, _ = json.Marshal(
			[]model.TaskError{
				{
					Severity:    "Error",
					Description: m.Error,
				},
			})
	}
	m := db.Migrator()
	err = m.DropColumn(&model.Task{}, "Error")
	return
}

func (r Migration) taskReportError(db *gorm.DB) (err error) {
	type TaskReport struct {
		model.TaskReport
		Error string
	}
	var list []TaskReport
	err = db.Find(&TaskReport{}, &list).Error
	if err != nil {
		return
	}
	for i := range list {
		m := &list[i]
		if m.Error == "" {
			continue
		}
		m.Errors, _ = json.Marshal(
			[]model.TaskError{
				{
					Severity:    "Error",
					Description: m.Error,
				},
			})
	}
	m := db.Migrator()
	err = m.DropColumn(&model.TaskReport{}, "Error")
	return
}
