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
	Load(client k8s.Client) (err error)
	Find(name string) (found bool)
	Match(capability string) (names []string, err error)
}

type BaseResolver struct {
}

type AddonResolver struct {
	BaseResolver
	addons map[string]*crd.Addon
}

func (r *AddonResolver) Load(client k8s.Client) (err error) {
	addons := crd.AddonList{}
	err = client.List(
		context.TODO(),
		&addons,
		k8s.InNamespace(Settings.Hub.Namespace))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	r.addons = make(map[string]*crd.Addon)
	for i := range addons.Items {
		addon := &addons.Items[i]
		r.addons[addon.Name] = addon
	}
	return
}

func (r *AddonResolver) Find(name string) (found bool) {
	_, found = r.addons[name]
	return
}

func (r *AddonResolver) Match(capability string) (names []string, err error) {
	for _, addon := range r.addons {
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
	extensions map[string]*crd.Extension
	addon      string
}

func (r *ExtensionResolver) Find(name string) (found bool) {
	_, found = r.extensions[name]
	return
}

func (r *ExtensionResolver) Load(client k8s.Client) (err error) {
	extensions := crd.ExtensionList{}
	err = client.List(
		context.TODO(),
		&extensions,
		k8s.InNamespace(Settings.Hub.Namespace))
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	r.extensions = make(map[string]*crd.Extension)
	for i := range extensions.Items {
		extension := &extensions.Items[i]
		if r.addon == extension.Spec.Addon {
			r.extensions[extension.Name] = extension
		}
	}
	return
}

func (r *ExtensionResolver) Match(capability string) (names []string, err error) {
	for _, extension := range r.extensions {
		if extension.Spec.Capability == capability {
			names = append(
				names,
				extension.Name)
		}
	}
	return
}
