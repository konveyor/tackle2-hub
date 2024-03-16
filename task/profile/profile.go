package profile

import (
	"encoding/json"
	"fmt"

	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/model"
	"gorm.io/gorm"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

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

func New(tp *crd.TaskProfile) (p *Profile) {
	p = &Profile{}
	p.TaskProfileSpec = tp.Spec
	return
}

type Profile struct {
	crd.TaskProfileSpec
}

func (p *Profile) Apply(db *gorm.DB, client k8s.Client, task *model.Task) (err error) {
	err = p.setAddon(db, client, task)
	if err != nil {
		return
	}
	err = p.setComponent(db, client, task)
	if err != nil {
		return
	}
	return
}

func (p *Profile) setAddon(db *gorm.DB, client k8s.Client, task *model.Task) (err error) {
	for i := range p.Addon {
		var selector Selector
		var matched []string
		resolver := &AddonResolver{}
		resolver.client = client
		selector, err = NewSelector(p.Addon[i], resolver)
		if err != nil {
			return
		}
		matched, err = selector.Match(db, task)
		if err != nil {
			return
		}
		if len(matched) == 0 {
			err = &NotResolved{Kind: "addon"}
			return
		}
		task.Addon = matched[0]
	}
	return
}

func (p *Profile) setComponent(db *gorm.DB, client k8s.Client, task *model.Task) (err error) {
	for i := range p.Component {
		var selector Selector
		var matched []string
		resolver := &ComponentResolver{}
		resolver.client = client
		selector, err = NewSelector(p.Component[i], resolver)
		if err != nil {
			return
		}
		matched, err = selector.Match(db, task)
		if err != nil {
			return
		}
		task.Components, _ = json.Marshal(matched)
	}
	return
}
