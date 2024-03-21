package task

import (
	"context"

	liberr "github.com/jortel/go-utils/error"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

// Resolver used to resolve names and categories.
type Resolver interface {
	// Load resources.
	Load(client k8s.Client) (err error)
	// Find returns true when the named resource exists.
	Find(name string) (found bool)
	// Match returns the resources that provide the capability.
	Match(capability string) (names []string, err error)
}

// BaseResolver -
type BaseResolver struct {
}

// AddonResolver resolves addons.
type AddonResolver struct {
	BaseResolver
	addons map[string]*crd.Addon
	task   string
}

// Load addons.
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
		if addon.Spec.Task == r.task {
			r.addons[addon.Name] = addon
		}
	}
	return
}

// Find returns true when the addon exists.
func (r *AddonResolver) Find(name string) (found bool) {
	_, found = r.addons[name]
	return
}

// Match returns the addons that provide the capability.
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

// ExtensionResolver resolves extensions.
type ExtensionResolver struct {
	BaseResolver
	extensions map[string]*crd.Extension
	addon      string
}

// Load extensions compatible with the addon.
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

// Find returns true when the extension exists.
func (r *ExtensionResolver) Find(name string) (found bool) {
	_, found = r.extensions[name]
	return
}

// Match returns the extensions that provide the capability.
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
