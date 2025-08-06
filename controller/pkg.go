package controller

import (
	"github.com/konveyor/tackle2-hub/controller/addon"
	"github.com/konveyor/tackle2-hub/controller/extension"
	"github.com/konveyor/tackle2-hub/controller/schema"
	"github.com/konveyor/tackle2-hub/controller/task"
	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Add the controller.
func Add(mgr manager.Manager, db *gorm.DB) (err error) {
	err = addon.Add(mgr, db)
	if err != nil {
		return
	}
	err = extension.Add(mgr, db)
	if err != nil {
		return
	}
	err = task.Add(mgr, db)
	if err != nil {
		return
	}
	err = schema.Add(mgr, db)
	if err != nil {
		return
	}
	return
}
