package task

import (
	"context"
	"sync"

	liberr "github.com/jortel/go-utils/error"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
	"github.com/konveyor/tackle2-hub/internal/k8s/simulator"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/labels"
	k8s "sigs.k8s.io/controller-runtime/pkg/client"
)

// NewCluster returns a configured cluster.
func NewCluster(client k8s.Client) (cluster Cluster) {
	if Settings.Hub.Task.Simulated {
		client = simulator.New()
	}
	cluster = Cluster{Client: client}
	return
}

// Cluster provides cached cluster resources.
// Maps must NOT be accessed directly.
type Cluster struct {
	k8s.Client
	mutex      sync.RWMutex
	tackle     *crd.Tackle
	addons     map[string]*crd.Addon
	extensions map[string]*crd.Extension
	tasks      map[string]*crd.Task
	quotas     map[string]*core.ResourceQuota
	pods       struct {
		other map[string]*core.Pod
		tasks map[string]*core.Pod
	}
}

// Refresh the cache.
func (k *Cluster) Refresh() (err error) {
	k.mutex.Lock()
	defer k.mutex.Unlock()
	if !Settings.Hub.Task.Enabled {
		k.tackle = &crd.Tackle{}
		k.addons = make(map[string]*crd.Addon)
		k.extensions = make(map[string]*crd.Extension)
		k.tasks = make(map[string]*crd.Task)
		k.pods.other = make(map[string]*core.Pod)
		k.pods.tasks = make(map[string]*core.Pod)
		k.quotas = make(map[string]*core.ResourceQuota)
		return
	}
	err = k.getTackle()
	if err != nil {
		return
	}
	err = k.getAddons()
	if err != nil {
		return
	}
	err = k.getExtensions()
	if err != nil {
		return
	}
	err = k.getTasks()
	if err != nil {
		return
	}
	err = k.getPods()
	if err != nil {
		return
	}
	err = k.getQuotas()
	if err != nil {
		return
	}
	return
}

// Tackle returns the tackle resource.
func (k *Cluster) Tackle() (r *crd.Tackle) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	r = k.tackle
	return
}

// Addon returns an addon my name.
func (k *Cluster) Addon(name string) (r *crd.Addon, found bool) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	r, found = k.addons[name]
	return
}

// Addons returns an addon my name.
func (k *Cluster) Addons() (list []*crd.Addon) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.addons {
		list = append(list, r)
	}
	return
}

// Extension returns an extension by name.
func (k *Cluster) Extension(name string) (r *crd.Extension, found bool) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	r, found = k.extensions[name]
	return
}

// Extensions returns an extension my name.
func (k *Cluster) Extensions() (list []*crd.Extension) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.extensions {
		list = append(list, r)
	}
	return
}

// FindExtensions returns extensions by name.
func (k *Cluster) FindExtensions(names []string) (list []*crd.Extension, err error) {
	for _, name := range names {
		r, found := k.extensions[name]
		if !found {
			err = &ExtensionNotFound{name}
			return
		}
		list = append(list, r)
	}
	return
}

// Task returns a task by name.
func (k *Cluster) Task(name string) (r *crd.Task, found bool) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	r, found = k.tasks[name]
	return
}

// Pod returns a pod by name.
func (k *Cluster) Pod(name string) (r *core.Pod, found bool) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	r, found = k.pods.tasks[name]
	return
}

// OtherPods returns a list of non-task pods.
func (k *Cluster) OtherPods() (list []*core.Pod) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.pods.other {
		list = append(list, r)
	}
	return
}

// TaskPods returns a list of task pods.
func (k *Cluster) TaskPods() (list []*core.Pod) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.pods.tasks {
		list = append(list, r)
	}
	return
}

// TaskPodsScheduled returns a list of task pods scheduled.
func (k *Cluster) TaskPodsScheduled() (list []*core.Pod) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.pods.tasks {
		if r.Status.Phase == core.PodFailed || r.Status.Phase == core.PodSucceeded {
			continue
		}
		list = append(list, r)
	}
	return
}

// Quotas returns quotas.
func (k *Cluster) Quotas() (list []*core.ResourceQuota) {
	k.mutex.RLock()
	defer k.mutex.RUnlock()
	for _, r := range k.quotas {
		list = append(list, r)
	}
	return
}

// PodQuota returns the most restricted pod quota.
func (k *Cluster) PodQuota() (quota int, found bool) {
	for _, r := range k.Quotas() {
		var qty resource.Quantity
		qty, found = r.Spec.Hard[core.ResourcePods]
		if found {
			n := int(qty.Value())
			if quota == 0 || quota > n {
				quota = n
			}
		}
	}
	return
}

// getTackle
func (k *Cluster) getTackle() (err error) {
	options := &k8s.ListOptions{Namespace: Settings.Namespace}
	list := crd.TackleList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		k.tackle = r
		return
	}
	err = liberr.New("Tackle CR not found.")
	return
}

// getAddons
func (k *Cluster) getAddons() (err error) {
	k.addons = make(map[string]*crd.Addon)
	options := &k8s.ListOptions{Namespace: Settings.Namespace}
	list := crd.AddonList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		k.addons[r.Name] = r
		if !r.Reconciled() {
			err = &NotReconciled{
				Kind: r.Kind,
				Name: r.Name,
			}
			return
		}
	}
	return
}

// getExtensions
func (k *Cluster) getExtensions() (err error) {
	k.extensions = make(map[string]*crd.Extension)
	options := &k8s.ListOptions{Namespace: Settings.Namespace}
	list := crd.ExtensionList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		k.extensions[r.Name] = r
	}
	return
}

// getTasks kinds.
func (k *Cluster) getTasks() (err error) {
	k.tasks = make(map[string]*crd.Task)
	options := &k8s.ListOptions{Namespace: Settings.Namespace}
	list := crd.TaskList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		k.tasks[r.Name] = r
	}
	return
}

// getPods
func (k *Cluster) getPods() (err error) {
	k.pods.other = make(map[string]*core.Pod)
	k.pods.tasks = make(map[string]*core.Pod)
	selector := labels.NewSelector()
	options := &k8s.ListOptions{
		Namespace:     Settings.Namespace,
		LabelSelector: selector,
	}
	list := core.PodList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		if _, found := r.Labels[TaskLabel]; found {
			k.pods.tasks[r.Name] = r
		} else {
			k.pods.other[r.Name] = r
		}

	}
	return
}

// getQuotas
func (k *Cluster) getQuotas() (err error) {
	k.quotas = make(map[string]*core.ResourceQuota)
	options := &k8s.ListOptions{Namespace: Settings.Namespace}
	list := core.ResourceQuotaList{}
	err = k.List(
		context.TODO(),
		&list,
		options)
	if err != nil {
		err = liberr.Wrap(err)
		return
	}
	for i := range list.Items {
		r := &list.Items[i]
		k.quotas[r.Name] = r
	}
	return
}
