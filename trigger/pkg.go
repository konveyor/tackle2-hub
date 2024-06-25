package trigger

import (
	"context"

	liberr "github.com/jortel/go-utils/error"
	crd "github.com/konveyor/tackle2-hub/k8s/api/tackle/v1alpha2"
	"github.com/konveyor/tackle2-hub/settings"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes/scheme"
	k8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

var (
	Settings = &settings.Settings
)

// Trigger supports actions triggered by model changes.
type Trigger struct {
}

// FindTasks returns tasks with the specified label.
func (r *Trigger) FindTasks(label string) (matched []*crd.Task, err error) {
	cfg, _ := config.GetConfig()
	client, err := k8sclient.New(
		cfg,
		k8sclient.Options{
			Scheme: scheme.Scheme,
		})
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
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
	err = client.List(
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
