package profile

import (
	"context"
	"fmt"

	liberr "github.com/jortel/go-utils/error"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/settings"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

var Settings = &settings.Settings

type NotResolved struct {
	Kind string
	Name string
}

func (e *NotResolved) Error() (s string) {
	return fmt.Sprintf("%s: '%s' not-resolved.", e.Kind, e.Name)
}

func (e *NotResolved) Is(err error) (matched bool) {
	_, matched = err.(*NotResolved)
	return
}

type Resolver interface {
	Match(capability string) (names []string, err error)
}

type BaseResolver struct {
	client k8s.Client
}

type AddonResolver struct {
	BaseResolver
}

func (r *AddonResolver) Match(capability string) (names []string, err error) {
	addons := crd.AddonList{}
	err = r.client.List(
		context.TODO(),
		&addons,
		k8s.InNamespace(Settings.Hub.Namespace))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, addon := range addons.Items {
		if addon.Spec.Capability == capability {
			names = append(
				names,
				addon.Name)
		}
	}
	return
}

type ExtensionResolver struct {
	BaseResolver
}

func (r *ExtensionResolver) Match(capability string) (names []string, err error) {
	extensions := crd.ExtensionList{}
	err = r.client.List(
		context.TODO(),
		&extensions,
		k8s.InNamespace(Settings.Hub.Namespace))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for _, extension := range extensions.Items {
		if extension.Spec.Capability == capability {
			names = append(
				names,
				extension.Name)
		}
	}
	return
}
