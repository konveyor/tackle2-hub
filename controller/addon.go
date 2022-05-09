package controller

import (
	"context"
	libcnd "github.com/konveyor/controller/pkg/condition"
	liberr "github.com/konveyor/controller/pkg/error"
	"github.com/konveyor/controller/pkg/logging"
	api "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"github.com/konveyor/tackle2-hub/settings"
	"gorm.io/gorm"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apiserver/pkg/storage/names"
	"k8s.io/client-go/tools/record"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const (
	Name = "addon"
)

//
// Package logger.
var log = logging.WithName(Name)

//
// Settings defines applcation settings.
var Settings = &settings.Settings

//
// Add the controller.
func Add(mgr manager.Manager, db *gorm.DB, adminChanged chan int) error {
	reconciler := &Reconciler{
		EventRecorder: mgr.GetRecorder(Name),
		Client:        mgr.GetClient(),
		AdminChanged:  adminChanged,
		Log:           log,
		DB:            db,
	}
	cnt, err := controller.New(
		Name,
		mgr,
		controller.Options{
			Reconciler: reconciler,
		})
	if err != nil {
		log.Trace(err)
		return err
	}
	// Primary CR.
	err = cnt.Watch(
		&source.Kind{Type: &api.Addon{}},
		&handler.EnqueueRequestForObject{})
	if err != nil {
		log.Trace(err)
		return err
	}

	return nil
}

//
// Reconciler reconciles addon CRs.
type Reconciler struct {
	record.EventRecorder
	k8s.Client
	DB           *gorm.DB
	AdminChanged chan int
	Log          *logging.Logger
}

//
// Reconcile a Addon CR.
// Note: Must not a pointer receiver to ensure that the
// logger and other state is not shared.
func (r Reconciler) Reconcile(request reconcile.Request) (result reconcile.Result, err error) {
	r.Log = logging.WithName(
		names.SimpleNameGenerator.GenerateName(Name+"|"),
		"addon",
		request)

	// Fetch the CR.
	addon := &api.Addon{}
	err = r.Get(context.TODO(), request.NamespacedName, addon)
	if err != nil {
		if k8serr.IsNotFound(err) {
			r.Log.Info("Addon deleted.", "name", request)
			_ = r.addonDeleted(request.Name)
			err = nil
		}
		return
	}

	//
	// changed.
	err = r.addonChanged(addon)
	if err != nil {
		return
	}

	// Begin staging conditions.
	addon.Status.BeginStagingConditions()

	// Ready condition.
	if !addon.Status.HasBlockerCondition() {
		addon.Status.SetCondition(libcnd.Condition{
			Type:     libcnd.Ready,
			Status:   "True",
			Category: "Required",
			Message:  "The addon is ready.",
		})
	}

	// End staging conditions.
	addon.Status.EndStagingConditions()

	// Apply changes.
	addon.Status.ObservedGeneration = addon.Generation
	err = r.Status().Update(context.TODO(), addon)
	if err != nil {
		return
	}

	// Done
	return
}

//
// addonChanged an addon has been created/updated.
// After the "admin" addon has reconciled, the volumes need
// be re-populated and the volume manager notified.
func (r *Reconciler) addonChanged(addon *api.Addon) (err error) {
	if addon.Name != "admin" {
		return
	}
	err = r.addonDeleted(addon.Name)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, mount := range addon.Spec.Mounts {
		m := &model.Volume{}
		m.Name = mount.Name
		err = r.DB.Create(m).Error
		if err != nil {
			err = liberr.Wrap(err)
			return
		}
	}
	func() { // send
		defer func() {
			recover()
		}()
		select {
		case r.AdminChanged <- 1:
		default:
		}
	}()

	return
}

//
// addonDeleted an addon has been deleted.
// The volumes need to be deleted.
func (r *Reconciler) addonDeleted(name string) (err error) {
	if name != "admin" {
		return
	}
	err = r.DB.Delete(&model.Volume{}, "id > 0").Error
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	return
}
