package simulator

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"

	core "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/konveyor/tackle2-hub/internal/k8s"
	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
)

// Client simulates a Kubernetes cluster for testing.
// Embeds k8s.FakeClient to get default no-op implementations.
type Client struct {
	k8s.FakeClient
	mutex      sync.RWMutex
	pods       map[string]*podEntry
	secrets    map[string]*core.Secret
	quotas     map[string]*core.ResourceQuota
	tackle     map[string]*crd.Tackle
	addons     map[string]*crd.Addon
	extensions map[string]*crd.Extension
	tasks      map[string]*crd.Task
	// Simulation parameters
	pendingDuration    time.Duration // How long pod stays in Pending
	runningDuration    time.Duration // How long pod stays in Running before Succeeded
	failureProbability float64       // Probability pod will fail instead of succeed
}

// podEntry tracks a pod and its creation time for state simulation.
type podEntry struct {
	pod       *core.Pod
	createdAt time.Time
}

// New creates a new simulator client with default timing and operator-installed resources.
func New() *Client {
	c := &Client{
		pods:               make(map[string]*podEntry),
		secrets:            make(map[string]*core.Secret),
		quotas:             make(map[string]*core.ResourceQuota),
		tackle:             make(map[string]*crd.Tackle),
		addons:             make(map[string]*crd.Addon),
		extensions:         make(map[string]*crd.Extension),
		tasks:              make(map[string]*crd.Task),
		pendingDuration:    10 * time.Second,
		runningDuration:    20 * time.Second,
		failureProbability: 0.0, // No failures by default
	}
	c.seed("addon.yaml", &crd.Addon{})
	c.seed("extension.yaml", &crd.Extension{})
	c.seed("task.yaml", &crd.Task{})
	return c
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
func (c *Client) Get(_ context.Context, key client.ObjectKey, obj client.Object, _ ...client.GetOption) (err error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	switch r := obj.(type) {
	case *core.Pod:
		entry, found := c.pods[key.Name]
		if !found {
			err = &meta.NoKindMatchError{}
			return
		}
		c.updatePodState(entry)
		*r = *entry.pod.DeepCopy()
	case *core.Secret:
		secret, found := c.secrets[key.Name]
		if !found {
			err = &meta.NoKindMatchError{}
			return
		}
		*r = *secret.DeepCopy()
	case *core.ResourceQuota:
		quota, found := c.quotas[key.Name]
		if !found {
			err = &meta.NoKindMatchError{}
			return
		}
		*r = *quota.DeepCopy()
	case *crd.Tackle:
		tackle, found := c.tackle[key.Name]
		if !found {
			err = &meta.NoKindMatchError{}
			return
		}
		*r = *tackle.DeepCopy()
	case *crd.Addon:
		addon, found := c.addons[key.Name]
		if !found {
			err = &meta.NoKindMatchError{}
			return
		}
		*r = *addon.DeepCopy()
	case *crd.Extension:
		extension, found := c.extensions[key.Name]
		if !found {
			err = &meta.NoKindMatchError{}
			return
		}
		*r = *extension.DeepCopy()
	case *crd.Task:
		task, found := c.tasks[key.Name]
		if !found {
			err = &meta.NoKindMatchError{}
			return
		}
		*r = *task.DeepCopy()
	default:
		err = &meta.NoKindMatchError{}
	}
	return
}

// List retrieves a list of resources.
func (c *Client) List(_ context.Context, list client.ObjectList, _ ...client.ListOption) (err error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	switch l := list.(type) {
	case *core.PodList:
		l.Items = make([]core.Pod, 0, len(c.pods))
		for _, entry := range c.pods {
			c.updatePodState(entry)
			l.Items = append(l.Items, *entry.pod.DeepCopy())
		}
	case *core.SecretList:
		l.Items = make([]core.Secret, 0, len(c.secrets))
		for _, secret := range c.secrets {
			l.Items = append(l.Items, *secret.DeepCopy())
		}
	case *core.ResourceQuotaList:
		l.Items = make([]core.ResourceQuota, 0, len(c.quotas))
		for _, quota := range c.quotas {
			l.Items = append(l.Items, *quota.DeepCopy())
		}
	case *crd.TackleList:
		l.Items = make([]crd.Tackle, 0, len(c.tackle))
		for _, tackle := range c.tackle {
			l.Items = append(l.Items, *tackle.DeepCopy())
		}
	case *crd.AddonList:
		l.Items = make([]crd.Addon, 0, len(c.addons))
		for _, addon := range c.addons {
			l.Items = append(l.Items, *addon.DeepCopy())
		}
	case *crd.ExtensionList:
		l.Items = make([]crd.Extension, 0, len(c.extensions))
		for _, extension := range c.extensions {
			l.Items = append(l.Items, *extension.DeepCopy())
		}
	case *crd.TaskList:
		l.Items = make([]crd.Task, 0, len(c.tasks))
		for _, task := range c.tasks {
			l.Items = append(l.Items, *task.DeepCopy())
		}
	default:
		err = &meta.NoKindMatchError{}
	}
	return
}

// Create creates a new resource.
func (c *Client) Create(_ context.Context, obj client.Object, _ ...client.CreateOption) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	switch r := obj.(type) {
	case *core.Pod:
		if _, found := c.pods[r.Name]; found {
			err = ConflictError{
				Kind: "Pod",
				Name: r.Name,
			}
			return
		}
		pod := r.DeepCopy()
		pod.Status.Phase = core.PodPending
		pod.Status.Conditions = []core.PodCondition{
			{
				Type:               core.PodScheduled,
				Status:             core.ConditionTrue,
				LastTransitionTime: meta_v1.Now(),
			},
		}
		if pod.UID == "" {
			pod.UID = newUID()
		}
		c.pods[pod.Name] = &podEntry{
			pod:       pod,
			createdAt: time.Now(),
		}
		*r = *pod
	case *core.Secret:
		if _, found := c.secrets[r.Name]; found {
			err = ConflictError{
				Kind: "Secret",
				Name: r.Name,
			}
			return
		}
		secret := r.DeepCopy()
		// Generate UID if not set
		if secret.UID == "" {
			secret.UID = newUID()
		}
		c.secrets[secret.Name] = secret
		*r = *secret
	case *core.ResourceQuota:
		if _, found := c.quotas[r.Name]; found {
			err = ConflictError{
				Kind: "Quota",
				Name: r.Name,
			}
			return
		}
		c.quotas[r.Name] = r.DeepCopy()
	case *crd.Tackle:
		if _, found := c.tackle[r.Name]; found {
			err = ConflictError{
				Kind: "Tackle",
				Name: r.Name,
			}
			return
		}
		c.tackle[r.Name] = r.DeepCopy()
	case *crd.Addon:
		if _, found := c.addons[r.Name]; found {
			err = ConflictError{
				Kind: "Addon",
				Name: r.Name,
			}
			return
		}
		c.addons[r.Name] = r.DeepCopy()
	case *crd.Extension:
		if _, found := c.extensions[r.Name]; found {
			err = ConflictError{
				Kind: "Extension",
				Name: r.Name,
			}
			return
		}
		c.extensions[r.Name] = r.DeepCopy()
	case *crd.Task:
		if _, found := c.tasks[r.Name]; found {
			err = ConflictError{
				Kind: "Task",
				Name: r.Name,
			}
			return
		}
		c.tasks[r.Name] = r.DeepCopy()
	default:
		err = fmt.Errorf("unsupported resource type: %T", obj)
	}
	return
}

// Delete removes a resource.
func (c *Client) Delete(_ context.Context, obj client.Object, _ ...client.DeleteOption) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	switch r := obj.(type) {
	case *core.Pod:
		delete(c.pods, r.Name)

	case *core.Secret:
		delete(c.secrets, r.Name)

	case *core.ResourceQuota:
		delete(c.quotas, r.Name)

	case *crd.Tackle:
		delete(c.tackle, r.Name)

	case *crd.Addon:
		delete(c.addons, r.Name)

	case *crd.Extension:
		delete(c.extensions, r.Name)

	case *crd.Task:
		delete(c.tasks, r.Name)

	default:
		err = fmt.Errorf("unsupported resource type: %T", obj)
	}
	return
}

// DeleteAllOf deletes all resources matching the criteria.
func (c *Client) DeleteAllOf(_ context.Context, obj client.Object, _ ...client.DeleteAllOfOption) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	switch obj.(type) {
	case *core.Pod:
		c.pods = make(map[string]*podEntry)
	case *core.Secret:
		c.secrets = make(map[string]*core.Secret)
	case *core.ResourceQuota:
		c.quotas = make(map[string]*core.ResourceQuota)
	case *crd.Tackle:
		c.tackle = make(map[string]*crd.Tackle)
	case *crd.Addon:
		c.addons = make(map[string]*crd.Addon)
	case *crd.Extension:
		c.extensions = make(map[string]*crd.Extension)
	case *crd.Task:
		c.tasks = make(map[string]*crd.Task)
	default:
		err = fmt.Errorf("unsupported resource type: %T", obj)
	}
	return
}

// Update updates a resource.
func (c *Client) Update(_ context.Context, obj client.Object, _ ...client.UpdateOption) (err error) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	switch r := obj.(type) {
	case *core.Pod:
		entry, found := c.pods[r.Name]
		if !found {
			err = fmt.Errorf("pod %s not found", r.Name)
			return
		}
		// Preserve creation time and update pod
		entry.pod = r.DeepCopy()

	case *core.Secret:
		if _, found := c.secrets[r.Name]; !found {
			err = fmt.Errorf("secret %s not found", r.Name)
			return
		}
		c.secrets[r.Name] = r.DeepCopy()

	case *core.ResourceQuota:
		if _, found := c.quotas[r.Name]; !found {
			err = fmt.Errorf("quota %s not found", r.Name)
			return
		}
		c.quotas[r.Name] = r.DeepCopy()

	case *crd.Tackle:
		if _, found := c.tackle[r.Name]; !found {
			err = fmt.Errorf("tackle %s not found", r.Name)
			return
		}
		c.tackle[r.Name] = r.DeepCopy()

	case *crd.Addon:
		if _, found := c.addons[r.Name]; !found {
			err = fmt.Errorf("addon %s not found", r.Name)
			return
		}
		c.addons[r.Name] = r.DeepCopy()

	case *crd.Extension:
		if _, found := c.extensions[r.Name]; !found {
			err = fmt.Errorf("extension %s not found", r.Name)
			return
		}
		c.extensions[r.Name] = r.DeepCopy()

	case *crd.Task:
		if _, found := c.tasks[r.Name]; !found {
			err = fmt.Errorf("task %s not found", r.Name)
			return
		}
		c.tasks[r.Name] = r.DeepCopy()

	default:
		err = fmt.Errorf("unsupported resource type: %T", obj)
	}
	return
}

// Patch patches a resource.
func (c *Client) Patch(ctx context.Context, obj client.Object, _ client.Patch, _ ...client.PatchOption) (err error) {
	err = c.Update(ctx, obj)
	return
}

// updatePodState updates a pod's state based on time elapsed since creation.
// This method should be called while holding at least a read lock.
func (c *Client) updatePodState(entry *podEntry) {
	elapsed := time.Since(entry.createdAt)
	pod := entry.pod
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
				LastTransitionTime: meta_v1.Now(),
			},
		}
		for i, container := range pod.Spec.Containers {
			containerStatuses[i] = core.ContainerStatus{
				Name:  container.Name,
				Ready: true,
				State: core.ContainerState{
					Running: &core.ContainerStateRunning{
						StartedAt: meta_v1.NewTime(entry.createdAt.Add(c.pendingDuration)),
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
					FinishedAt: meta_v1.Now(),
				},
			},
		}
	}
	pod.Status.ContainerStatuses = containerStatuses
}

// podFailed determines if a pod should fail based on configured probability.
// For deterministic behavior, use pod name hash
// In production, you might want to use random based on probability
// Simple hash-based deterministic failure
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

func (c *Client) seed(path string, r client.Object) {
	ctx := context.TODO()
	path = filepath.Join(dataDir(), path)
	b, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}
	content := strings.Split(string(b), "\n---\n")
	for _, d := range content {
		err = yaml.Unmarshal([]byte(d), r)
		err = c.Create(ctx, r)
		if err != nil {
			panic(err)
		}
	}
	return
}

// dataDir returns the path to the data directory.
func dataDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "data")
}

// newUID generates a simple UID for resources.
func newUID() types.UID {
	return types.UID(fmt.Sprintf("%d", time.Now().UnixNano()))
}

type ConflictError struct {
	Kind string
	Name string
}

func (e ConflictError) Error() string {
	return fmt.Sprintf("(%s) %s already exists.", e.Kind, e.Name)
}
