package controller

import (
	"context"

	"github.com/go-logr/logr"
	logr2 "github.com/jortel/go-utils/logr"
	api "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha2"
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

// Package logger.
var log = logr2.WithName(Name)

// Settings defines applcation settings.
var Settings = &settings.Settings

// Add the controller.
func Add(mgr manager.Manager, db *gorm.DB) error {
	reconciler := &Reconciler{
		Client: mgr.GetClient(),
		Log:    log,
		DB:     db,
	}
	cnt, err := controller.New(
		Name,
		mgr,
		controller.Options{
			Reconciler: reconciler,
		})
	if err != nil {
		log.Error(err, "")
		return err
	}
	// Primary CR.
	err = cnt.Watch(
		&source.Kind{Type: &api.Addon{}},
		&handler.EnqueueRequestForObject{})
	if err != nil {
		log.Error(err, "")
		return err
	}

	return nil
}

// Reconciler reconciles addon CRs.
type Reconciler struct {
	record.EventRecorder
	k8s.Client
	DB  *gorm.DB
	Log logr.Logger
}

// Reconcile a Addon CR.
// Note: Must not a pointer receiver to ensure that the
// logger and other state is not shared.
func (r Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result, err error) {
	r.Log = logr2.WithName(
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
	// migrate
	migrated, err := r.alpha2Migration(addon)
	if migrated || err != nil {
		return
	}
	// changed.
	err = r.addonChanged(addon)
	if err != nil {
		return
	}
	// Apply changes.
	addon.Status.ObservedGeneration = addon.Generation
	err = r.Status().Update(context.TODO(), addon)
	if err != nil {
		return
	}

	return
}

// addonChanged an addon has been created/updated.
func (r *Reconciler) addonChanged(addon *api.Addon) (err error) {
	return
}

// addonDeleted an addon has been deleted.
func (r *Reconciler) addonDeleted(name string) (err error) {
	return
}

// alpha2Migration migrates to alpha2.
func (r *Reconciler) alpha2Migration(addon *api.Addon) (migrated bool, err error) {
	if addon.Spec.Image != nil {
		if addon.Spec.Container.Image == "" {
			addon.Spec.Container.Image = *addon.Spec.Image
		}
		addon.Spec.Image = nil
		migrated = true
	}
	if addon.Spec.Resources != nil {
		if len(addon.Spec.Container.Resources.Limits) == 0 {
			addon.Spec.Container.Resources.Limits = (*addon.Spec.Resources).Limits
		}
		if len(addon.Spec.Container.Resources.Requests) == 0 {
			addon.Spec.Container.Resources.Requests = (*addon.Spec.Resources).Requests
		}
		addon.Spec.Resources = nil
		migrated = true
	}
	if addon.Spec.ImagePullPolicy != nil {
		if addon.Spec.Container.ImagePullPolicy == "" {
			addon.Spec.Container.ImagePullPolicy = *addon.Spec.ImagePullPolicy
		}
		addon.Spec.ImagePullPolicy = nil
		migrated = true
	}
	if migrated {
		err = r.Update(context.TODO(), addon)
		if err != nil {
			return
		}
	}
	return
}
