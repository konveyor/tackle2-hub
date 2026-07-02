package controller

import (
	"github.com/konveyor/tackle2-hub/internal/controller/addon"
	"github.com/konveyor/tackle2-hub/internal/controller/client"
	"github.com/konveyor/tackle2-hub/internal/controller/idp"
	"github.com/konveyor/tackle2-hub/internal/controller/ldap"
	"gorm.io/gorm"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// Add the controllers.
func Add(mgr manager.Manager, db *gorm.DB) (err error) {
	err = addon.Add(mgr, db)
	if err != nil {
		return
	}
	err = idp.Add(mgr, db)
	if err != nil {
		return
	}
	err = ldap.Add(mgr, db)
	if err != nil {
		return
	}
	err = client.Add(mgr, db)
	if err != nil {
		return
	}
	return
}
