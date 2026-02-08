package simulator

import (
	"context"
	"fmt"
	"time"

	"github.com/konveyor/tackle2-hub/internal/k8s/seed"
	core "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// FakeClient interface.
// Provided to support behavior adjustments for testing.
type FakeClient interface {
	client.Client
	Use(monitor PodMonitor) *Client
}

// Client simulates a Kubernetes cluster for testing.
// Wraps the fake client and adds pod lifecycle simulation.
type Client struct {
	client.Client
	podMonitor PodMonitor
}

// New creates a new simulator with seeded resources.
func New() *Client {
	b := fake.NewClientBuilder()
	fakeClient :=
		b.WithScheme(seed.Scheme()).
			WithObjects(seed.Resources()...).
			Build()
	return &Client{
		Client:     fakeClient,
		podMonitor: NewMonitor(5, 10),
	}
}

// Use pod monitor.
func (c *Client) Use(monitor PodMonitor) *Client {
	c.podMonitor = monitor
	return c
}

// Get retrieves a resource by key.
func (c *Client) Get(
	ctx context.Context,
	key client.ObjectKey,
	object client.Object,
	opts ...client.GetOption) (err error) {
	//
	err = c.Client.Get(ctx, key, object, opts...)
	if err != nil {
		return
	}
	switch r := object.(type) {
	case *core.Pod:
		pod := r
		err = c.updatePod(ctx, pod)
		if err != nil {
			return
		}
	default:
		//
	}

	return
}

// List retrieves a list of resources.
func (c *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (err error) {
	err = c.Client.List(ctx, list, opts...)
	if err != nil {
		return
	}
	switch r := list.(type) {
	case *core.PodList:
		for i := range r.Items {
			err = c.updatePod(ctx, &r.Items[i])
			if err != nil {
				return
			}
		}
	default:
		//
	}

	return
}

// Create creates a new resource.
func (c *Client) Create(ctx context.Context, object client.Object, opts ...client.CreateOption) (err error) {
	switch r := object.(type) {
	case *core.Pod:
		pod := r
		c.podCreated(pod)
	default:
		//
	}
	err = c.Client.Create(ctx, object, opts...)
	if err != nil {
		return
	}
	switch r := object.(type) {
	case *core.Pod:
		pod := r
		c.podMonitor.Created(pod)
	default:
		//
	}

	return
}

// Delete removes a resource.
func (c *Client) Delete(ctx context.Context, object client.Object, opts ...client.DeleteOption) (err error) {
	err = c.Client.Delete(ctx, object, opts...)
	if err != nil {
		return
	}
	switch r := object.(type) {
	case *core.Pod:
		pod := r
		c.podMonitor.Deleted(pod)
	default:
		//
	}
	return
}

// updatePod updates a pod's disposition.
func (c *Client) updatePod(ctx context.Context, pod *core.Pod) (err error) {
	current := pod.Status.Phase
	next := c.podMonitor.Next(pod)
	switch next {
	case core.PodPending:
		c.podPending(pod)
	case core.PodRunning:
		c.podRunning(pod)
	case core.PodSucceeded:
		c.podSucceeded(pod)
	case core.PodFailed:
		c.podFailed(pod)
	default:
		phase := field.Invalid(
			field.NewPath("status").
				Child("phase"),
			next,
			"invalid pod phase returned by monitor",
		)
		err = k8serr.NewInvalid(
			schema.GroupKind{
				Group: "",
				Kind:  "Pod"},
			pod.Name,
			field.ErrorList{
				phase,
			},
		)
		return
	}
	dirty := next != current
	if dirty {
		err = c.Update(ctx, pod)
	}
	return
}

// podCreated updates the pod to reflect a scheduled state.
func (c *Client) podCreated(pod *core.Pod) {
	pod.UID = newUID()
	pod.Status.Phase = core.PodPending
	pod.Status.Conditions = []core.PodCondition{
		{
			Type:               core.PodScheduled,
			Status:             core.ConditionTrue,
			LastTransitionTime: meta.Now(),
		},
	}
}

// podPending updates the pod to reflect a pending state.
func (c *Client) podPending(pod *core.Pod) {
	pod.Status.Phase = core.PodPending
	statuses := make(
		[]core.ContainerStatus,
		len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {
		statuses[i] = core.ContainerStatus{
			Name:  container.Name,
			Ready: false,
			State: core.ContainerState{
				Waiting: &core.ContainerStateWaiting{
					Reason:  "ContainerCreating",
					Message: "Container is being created.",
				},
			},
		}
	}
	pod.Status.ContainerStatuses = statuses
}

// podRunning updates the pod to reflect a running state.
func (c *Client) podRunning(pod *core.Pod) {
	pod.Status.Phase = core.PodRunning
	statuses := make(
		[]core.ContainerStatus,
		len(pod.Spec.Containers))
	pod.Status.Conditions = []core.PodCondition{
		{
			Type:               core.PodReady,
			Status:             core.ConditionTrue,
			LastTransitionTime: meta.Now(),
		},
	}
	for i, container := range pod.Spec.Containers {
		statuses[i] = core.ContainerStatus{
			Name:  container.Name,
			Ready: true,
			State: core.ContainerState{
				Running: &core.ContainerStateRunning{
					StartedAt: pod.CreationTimestamp,
				},
			},
		}
	}
	pod.Status.ContainerStatuses = statuses
}

// podSucceeded updates the pod to reflect a succeeded state.
func (c *Client) podSucceeded(pod *core.Pod) {
	pod.Status.Phase = core.PodSucceeded
	statuses := make(
		[]core.ContainerStatus,
		len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {
		exitCode := 0
		reason := "Completed"
		message := "Container completed."
		statuses[i] = core.ContainerStatus{
			Name:  container.Name,
			Ready: false,
			State: core.ContainerState{
				Terminated: &core.ContainerStateTerminated{
					ExitCode:   int32(exitCode),
					Reason:     reason,
					Message:    message,
					FinishedAt: meta.Now(),
				},
			},
		}
	}
	pod.Status.ContainerStatuses = statuses
}

// podFailed updates the pod to reflect a failed state.
func (c *Client) podFailed(pod *core.Pod) {
	pod.Status.Phase = core.PodFailed
	statuses := make(
		[]core.ContainerStatus,
		len(pod.Spec.Containers))
	for i, container := range pod.Spec.Containers {
		exitCode := 1
		reason := "Error"
		message := "Container failed."
		statuses[i] = core.ContainerStatus{
			Name:  container.Name,
			Ready: false,
			State: core.ContainerState{
				Terminated: &core.ContainerStateTerminated{
					ExitCode:   int32(exitCode),
					Reason:     reason,
					Message:    message,
					FinishedAt: meta.Now(),
				},
			},
		}
	}
	pod.Status.ContainerStatuses = statuses
}

// newUID generates a simple UID for resources.
func newUID() (u types.UID) {
	u = types.UID(
		fmt.Sprintf("%d", time.Now().UnixNano()))
	return
}
