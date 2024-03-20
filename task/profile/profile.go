package profile

import (
	"encoding/json"
	"errors"
	"fmt"

	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

// AddonNotSelected report that an addon has not been selected.
type AddonNotSelected struct {
}

func (e *AddonNotSelected) Error() (s string) {
	return fmt.Sprintf("Addon not selected.")
}

func (e *AddonNotSelected) Is(err error) (matched bool) {
	var addonNotSelected *AddonNotSelected
	matched = errors.As(err, &addonNotSelected)
	return
}

// New returns a profile.
func New(tp *crd.TaskProfile) (p *Profile) {
	p = &Profile{}
	p.TaskProfileSpec = tp.Spec
	return
}

// Profile defines a kind of task.
type Profile struct {
	crd.TaskProfileSpec
}

// Apply the profile to the task. Sets the task addon and extensions.
func (p *Profile) Apply(db *gorm.DB, client k8s.Client, task *model.Task) (err error) {
	err = p.setAddon(db, client, task)
	if err != nil {
		return
	}
	err = p.setExtensions(db, client, task)
	if err != nil {
		return
	}
	return
}

// setAddon sets the task addon.
func (p *Profile) setAddon(db *gorm.DB, client k8s.Client, task *model.Task) (err error) {
	selected := ""
	for i := range p.Addon {
		var selector Selector
		var matched []string
		resolver := &AddonResolver{}
		err = resolver.Load(client)
		if err != nil {
			return
		}
		selector, err = NewSelector(p.Addon[i], resolver)
		if err != nil {
			return
		}
		matched, err = selector.Match(db, task)
		if err != nil {
			return
		}
		selected = matched[0]
		break
	}
	if selected == "" {
		err = &AddonNotSelected{}
		return
	}
	task.Addon = selected
	return
}

// setExtensions sets the task extensions.
func (p *Profile) setExtensions(db *gorm.DB, client k8s.Client, task *model.Task) (err error) {
	names := make(map[string]int)
	for i := range p.Extension {
		var selector Selector
		var matched []string
		resolver := &ExtensionResolver{
			addon: task.Addon,
		}
		err = resolver.Load(client)
		if err != nil {
			return
		}
		selector, err = NewSelector(p.Extension[i], resolver)
		if err != nil {
			return
		}
		matched, err = selector.Match(db, task)
		if err != nil {
			return
		}
		for _, name := range matched {
			names[name] = 0
		}
	}
	extensions := make([]string, len(names))
	for name := range names {
		extensions = append(
			extensions,
			name)
	}
	task.Extensions, _ = json.Marshal(extensions)
	return
}
