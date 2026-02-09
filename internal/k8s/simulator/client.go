package simulator

import (
	"context"

	"github.com/google/uuid"
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
	Use(manager PodManager) *Client
}

// Client simulates a Kubernetes cluster for testing.
// Wraps the fake client and adds pod lifecycle simulation.
type Client struct {
	client.Client
	podManager PodManager
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
		podManager: NewManager(5, 10),
	}
}

// Use pod manager.
func (s *Client) Use(manager PodManager) *Client {
	s.podManager = manager
	return s
}

// Get retrieves a resource by key.
func (s *Client) Get(
	ctx context.Context,
	key client.ObjectKey,
	object client.Object,
	options ...client.GetOption) (err error) {
	//
	err = s.Client.Get(ctx, key, object, options...)
	if err != nil {
		return
	}
	switch r := object.(type) {
	case *core.Pod:
		pod := r
		err = s.updatePod(ctx, pod)
		if err != nil {
			return
		}
	default:
		//
	}

	return
}

// List retrieves a list of resources.
func (s *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (err error) {
	err = s.Client.List(ctx, list, opts...)
	if err != nil {
		return
	}
	switch r := list.(type) {
	case *core.PodList:
		for i := range r.Items {
			err = s.updatePod(ctx, &r.Items[i])
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
func (s *Client) Create(ctx context.Context, object client.Object, opts ...client.CreateOption) (err error) {
	switch r := object.(type) {
	case *core.Pod:
		pod := r
		s.podCreated(pod)
	default:
		//
	}
	err = s.Client.Create(ctx, object, opts...)
	if err != nil {
		return
	}
	switch r := object.(type) {
	case *core.Pod:
		pod := r
		s.podManager.Created(pod)
	default:
		//
	}

	return
}

// Delete removes a resource.
func (s *Client) Delete(ctx context.Context, object client.Object, opts ...client.DeleteOption) (err error) {
	err = s.Client.Delete(ctx, object, opts...)
	if err != nil {
		return
	}
	switch r := object.(type) {
	case *core.Pod:
		pod := r
		s.podManager.Deleted(pod)
	default:
		//
	}
	return
}

// updatePod updates a pod's disposition.
func (s *Client) updatePod(ctx context.Context, pod *core.Pod) (err error) {
	current := pod.Status.Phase
	next := s.podManager.Next(pod)
	if next == current ||
		current == core.PodSucceeded ||
		current == core.PodFailed {
		return
	}
	switch next {
	case core.PodPending:
		s.podPending(pod)
	case core.PodRunning:
		s.podRunning(pod)
	case core.PodSucceeded:
		s.podSucceeded(pod)
	case core.PodFailed:
		s.podFailed(pod)
	default:
		phase := field.Invalid(
			field.NewPath("status").
				Child("phase"),
			next,
			"invalid pod phase returned by manager",
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
	err = s.Update(ctx, pod)
	return
}

// podCreated updates the pod to reflect a scheduled state.
func (s *Client) podCreated(pod *core.Pod) {
	pod.UID = s.newUID()
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
func (s *Client) podPending(pod *core.Pod) {
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
func (s *Client) podRunning(pod *core.Pod) {
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
func (s *Client) podSucceeded(pod *core.Pod) {
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
func (s *Client) podFailed(pod *core.Pod) {
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

// newUID generates a UID for resources.
func (s *Client) newUID() (u types.UID) {
	n, _ := uuid.NewUUID()
	u = types.UID(n.String())
	return
}
