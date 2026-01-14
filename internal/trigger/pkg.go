package trigger

import (
	"context"

	liberr "github.com/jortel/go-utils/error"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	tasking "github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/settings"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	Settings = &settings.Settings
)

// Trigger supports actions triggered by model changes.
type Trigger struct {
	User        string
	TaskManager *tasking.Manager
	Client      k8sclient.Client
	DB          *gorm.DB
}

// FindTasks returns tasks with the specified label.
func (r *Trigger) FindTasks(label string) (matched []*crd.Task, err error) {
	selector := labels.NewSelector()
	req, _ := labels.NewRequirement(
		label,
		selection.Exists,
		[]string{})
	selector = selector.Add(*req)
	options := &k8sclient.ListOptions{
		Namespace:     Settings.Namespace,
		LabelSelector: selector,
	}
	list := crd.TaskList{}
	err = r.Client.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		t := &list.Items[i]
		matched = append(matched, t)
	}
	return
}
