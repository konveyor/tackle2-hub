package simulator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/konveyor/tackle2-hub/internal/k8s/seed"
	core "k8s.io/api/core/v1"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Client simulates a Kubernetes cluster for testing.
// Wraps the fake client and adds pod lifecycle simulation.
type Client struct {
	client.Client
	mutex     sync.RWMutex
	podTiming map[string]time.Time
	// Simulation parameters
	pendingDuration    time.Duration // How long pod stays in Pending
	runningDuration    time.Duration // How long pod stays in Running before Succeeded
	failureProbability float64       // Probability pod will fail instead of succeed
}

// New creates a new simulator client with default timing and operator-installed resources.
func New() *Client {
	fakeClient := fake.NewClientBuilder().
		WithScheme(seed.Scheme()).
		WithObjects(seed.Resources()...).
		Build()
	return &Client{
		Client:             fakeClient,
		podTiming:          make(map[string]time.Time),
		pendingDuration:    5 * time.Second,
		runningDuration:    10 * time.Second,
		failureProbability: 0.0,
	}
}

// WithTiming sets custom timing for pod state progression.
// Parameters are number of seconds.
func (c *Client) WithTiming(pending, running int) *Client {
	c.pendingDuration = time.Duration(pending) * time.Second
	c.runningDuration = time.Duration(running) * time.Second
	return c
}

// WithFailureProbability sets the probability of pod failure (0.0 - 1.0).
func (c *Client) WithFailureProbability(prob float64) *Client {
	c.failureProbability = prob
	return c
}

// Get retrieves a resource by key.
func (c *Client) Get(
	ctx context.Context,
	key client.ObjectKey,
	obj client.Object,
	opts ...client.GetOption) (err error) {
	//
	err = c.Client.Get(ctx, key, obj, opts...)
	if err != nil {
		return
	}
	pod, isPod := obj.(*core.Pod)
	if !isPod {
		return
	}
	c.mutex.RLock()
	createdAt, found := c.podTiming[key.String()]
	c.mutex.RUnlock()
	if found {
		c.updatePodState(pod, createdAt)
	}
	return
}

// List retrieves a list of resources.
func (c *Client) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) (err error) {
	err = c.Client.List(ctx, list, opts...)
	if err != nil {
		return
	}
	podList, isPodList := list.(*core.PodList)
	if !isPodList {
		return
	}
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	for i := range podList.Items {
		pod := &podList.Items[i]
		key := client.ObjectKeyFromObject(pod).String()
		if createdAt, found := c.podTiming[key]; found {
			c.updatePodState(pod, createdAt)
		}
	}
	return
}

// Create creates a new resource.
func (c *Client) Create(ctx context.Context, obj client.Object, opts ...client.CreateOption) (err error) {
	if pod, isPod := obj.(*core.Pod); isPod {
		pod.Status.Phase = core.PodPending
		pod.Status.Conditions = []core.PodCondition{
			{
				Type:               core.PodScheduled,
				Status:             core.ConditionTrue,
				LastTransitionTime: meta.Now(),
			},
		}
		if pod.UID == "" {
			pod.UID = newUID()
		}
	}
	err = c.Client.Create(ctx, obj, opts...)
	if err != nil {
		return
	}
	if _, isPod := obj.(*core.Pod); isPod {
		c.mutex.Lock()
		key := client.ObjectKeyFromObject(obj).String()
		c.podTiming[key] = time.Now()
		c.mutex.Unlock()
	}
	return
}

// Delete removes a resource.
func (c *Client) Delete(ctx context.Context, obj client.Object, opts ...client.DeleteOption) (err error) {
	if _, isPod := obj.(*core.Pod); isPod {
		c.mutex.Lock()
		key := client.ObjectKeyFromObject(obj).String()
		delete(c.podTiming, key)
		c.mutex.Unlock()
	}
	err = c.Client.Delete(ctx, obj, opts...)
	return
}

// updatePodState updates a pod's state based on time elapsed since creation.
func (c *Client) updatePodState(pod *core.Pod, createdAt time.Time) {
	elapsed := time.Since(createdAt)
	// Build container statuses dynamically based on pod spec.
	// This supports pods with multiple containers (e.g., main task + extensions).
	containerStatuses := make([]core.ContainerStatus, len(pod.Spec.Containers))
	// Pod is in Pending state
	if elapsed < c.pendingDuration {
		pod.Status.Phase = core.PodPending
		for i, container := range pod.Spec.Containers {
			containerStatuses[i] = core.ContainerStatus{
				Name:  container.Name,
				Ready: false,
				State: core.ContainerState{
					Waiting: &core.ContainerStateWaiting{
						Reason:  "ContainerCreating",
						Message: "Container is being created",
					},
				},
			}
		}
		pod.Status.ContainerStatuses = containerStatuses
		return
	}
	// Pod is in Running state
	if elapsed < c.pendingDuration+c.runningDuration {
		pod.Status.Phase = core.PodRunning
		pod.Status.Conditions = []core.PodCondition{
			{
				Type:               core.PodReady,
				Status:             core.ConditionTrue,
				LastTransitionTime: meta.Now(),
			},
		}
		for i, container := range pod.Spec.Containers {
			containerStatuses[i] = core.ContainerStatus{
				Name:  container.Name,
				Ready: true,
				State: core.ContainerState{
					Running: &core.ContainerStateRunning{
						StartedAt: meta.NewTime(createdAt.Add(c.pendingDuration)),
					},
				},
			}
		}
		pod.Status.ContainerStatuses = containerStatuses
		return
	}
	// Pod has completed - decide if succeeded or failed
	failed := c.podFailed(pod)
	if failed {
		pod.Status.Phase = core.PodFailed
	} else {
		pod.Status.Phase = core.PodSucceeded
	}
	for i, container := range pod.Spec.Containers {
		exitCode := 0
		reason := "Completed"
		message := "Container completed successfully"
		if failed {
			exitCode = 1
			reason = "Error"
			message = "Simulated failure"
		}
		containerStatuses[i] = core.ContainerStatus{
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
	pod.Status.ContainerStatuses = containerStatuses
}

// podFailed determines if a pod should fail based on configured probability.
func (c *Client) podFailed(pod *core.Pod) (failed bool) {
	if c.failureProbability == 0.0 {
		return
	}
	hash := 0
	for _, ch := range pod.Name {
		hash = (hash*31 + int(ch)) % 100
	}
	failed = float64(hash) < (c.failureProbability * 100)
	return
}

// newUID generates a simple UID for resources.
func newUID() (u types.UID) {
	u = types.UID(
		fmt.Sprintf("%d", time.Now().UnixNano()))
	return
}
