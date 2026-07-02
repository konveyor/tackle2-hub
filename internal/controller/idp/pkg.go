package idp

import (
	"context"
	"strings"
	"syscall"

	"github.com/go-logr/logr"
	logr2 "github.com/jortel/go-utils/logr"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"gorm.io/gorm"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Name = "idp"
)

var Log = logr2.WithName(Name)

type IdentityProvider = crd.IdentityProvider

// Add the controller.
func Add(mgr manager.Manager, db *gorm.DB) (err error) {
	reconciler := &Reconciler{
		seen:   make(map[string]bool),
		Client: mgr.GetClient(),
		Log:    Log,
		DB:     db,
	}
	cnt, err := controller.New(
		Name,
		mgr,
		controller.Options{
			Reconciler: reconciler,
		})
	if err != nil {
		Log.Error(err, "")
		return
	}
	err = cnt.Watch(
		&source.Kind{Type: &IdentityProvider{}},
		&handler.EnqueueRequestForObject{})
	if err != nil {
		Log.Error(err, "")
		return
	}

	return
}

// Reconciler reconciles idp CRs.
// The seen (map) is used to ensure resources are
// reconciled at least once at startup.
type Reconciler struct {
	record.EventRecorder
	k8s.Client
	DB   *gorm.DB
	Log  logr.Logger
	seen map[string]bool
}

// Reconcile a Addon CR.
// Note: Must not be a pointer receiver to ensure that the
// logger and other state is not shared.
func (r Reconciler) Reconcile(ctx context.Context, request reconcile.Request) (result reconcile.Result, err error) {
	r.Log = logr2.WithName(
		names.SimpleNameGenerator.GenerateName(Name+"|"),
		"idp",
		request)

	// Fetch the CR.
	idp := &IdentityProvider{}
	err = r.Get(context.TODO(), request.NamespacedName, idp)
	if err != nil {
		if k8serr.IsNotFound(err) {
			_ = r.deleted(request.Name)
			err = nil
		}
		return
	}
	_, found := r.seen[idp.Name]
	if found && idp.Reconciled() {
		return
	}
	// Changed
	err = r.changed(idp)
	if err != nil {
		return
	}
	// Ready condition.
	ready := r.ready(idp)
	idp.Status.Conditions = nil
	idp.Status.ObservedGeneration = idp.Generation
	idp.Status.Conditions = append(
		idp.Status.Conditions,
		ready)
	// Apply changes.
	err = r.Status().Update(context.TODO(), idp)
	if err != nil {
		return
	}

	r.seen[idp.Name] = true

	return
}

// ready returns the ready condition.
func (r *Reconciler) ready(idp *IdentityProvider) (ready v1.Condition) {
	ready = crd.Ready
	ready.LastTransitionTime = v1.Now()
	ready.ObservedGeneration = idp.Status.ObservedGeneration
	err := make([]string, 0)
	for i := range idp.Status.Conditions {
		cnd := &idp.Status.Conditions[i]
		if cnd.Type == crd.ValidationError {
			err = append(err, cnd.Message)
		}
	}
	if len(err) == 0 {
		ready.Status = v1.ConditionTrue
		ready.Reason = crd.Validated
		ready.Message = strings.Join(err, ";")
	} else {
		ready.Status = v1.ConditionFalse
		ready.Reason = crd.ValidationError
	}
	return
}

// changed an idp has been created/updated.
// When detected, the hub is restarted.
func (r *Reconciler) changed(p *IdentityProvider) (err error) {
	if !r.seen[p.Name] {
		return
	}
	Log.Info(
		"IdP added/changed.",
		"name",
		p.Name)
	r.hubRestart()
	return
}

// deleted an idp has been deleted.
// When detected, the hub is restarted.
func (r *Reconciler) deleted(name string) (err error) {
	Log.Info(
		"Idp deleted.",
		"name",
		name)
	r.hubRestart()
	return
}

// hubRestart restarts the hub.
func (r *Reconciler) hubRestart() {
	Log.Info("**** RESTARTING HUB *****")
	_ = syscall.Kill(
		syscall.Getpid(),
		syscall.SIGTERM)
}
