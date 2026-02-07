package simulator

import (
	"context"
	"time"

	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	crd "github.com/konveyor/tackle2-hub/internal/k8s/api/tackle/v1alpha1"
)

// Example demonstrates how to use the simulator for testing task execution.
func Example() {
	// Create a simulator with custom timing
	// Pods will spend 5 seconds in Pending, then 10 seconds in Running
	sim := New().WithTiming(5*time.Second, 10*time.Second)

	// Seed the simulator with required resources
	seedSimulator(sim)

	// Use the simulator as a client
	// In real usage, you would inject this into Manager or Cluster:
	//   manager.Client = sim
	//   cluster := &task.Cluster{Client: sim}

	// For demonstration, let's show pod lifecycle
	demonstratePodLifecycle(sim)
}

// seedSimulator populates the simulator with cluster-level resources.
// Note: Operator-installed resources (Addons, Tasks, Extensions) are
// automatically loaded by New().
func seedSimulator(sim *Client) {
	ctx := context.TODO()

	// Create Tackle CR
	tackle := &crd.Tackle{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "tackle",
			Namespace: "konveyor-tackle",
		},
		Spec: runtime.RawExtension{
			Raw: []byte("URL: http://localhost:8080"),
		},
	}
	sim.Create(ctx, tackle)

	// Add resource quota
	quota := &core.ResourceQuota{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "default-quota",
			Namespace: "konveyor-tackle",
		},
		Spec: core.ResourceQuotaSpec{
			Hard: core.ResourceList{
				core.ResourcePods: *resource.NewQuantity(10, resource.DecimalSI),
			},
		},
	}
	sim.Create(ctx, quota)
}

// demonstratePodLifecycle shows how pods progress through states.
func demonstratePodLifecycle(sim *Client) {
	ctx := context.TODO()

	// Create a task pod
	pod := &core.Pod{
		ObjectMeta: meta_v1.ObjectMeta{
			Name:      "task-1-abc123",
			Namespace: "konveyor-tackle",
			Labels: map[string]string{
				"task": "1",
			},
		},
		Spec: core.PodSpec{
			Containers: []core.Container{
				{
					Name:  "main",
					Image: "quay.io/konveyor/analyzer:latest",
				},
			},
			RestartPolicy: core.RestartPolicyNever,
		},
	}

	// Create the pod
	sim.Create(ctx, pod)

	// Check initial state (should be Pending)
	retrieved := &core.Pod{}
	sim.Get(ctx, client.ObjectKey{Name: pod.Name, Namespace: pod.Namespace}, retrieved)
	// retrieved.Status.Phase == core.PodPending

	// Wait and check Running state
	time.Sleep(6 * time.Second)
	sim.Get(ctx, client.ObjectKey{Name: pod.Name, Namespace: pod.Namespace}, retrieved)
	// retrieved.Status.Phase == core.PodRunning
	// retrieved.Status.ContainerStatuses[0].Ready == true

	// Wait and check Succeeded state
	time.Sleep(10 * time.Second)
	sim.Get(ctx, client.ObjectKey{Name: pod.Name, Namespace: pod.Namespace}, retrieved)
	// retrieved.Status.Phase == core.PodSucceeded
	// retrieved.Status.ContainerStatuses[0].State.Terminated.ExitCode == 0
}

// ExampleWithManager demonstrates integrating the simulator with a task Manager.
/*
func ExampleWithManager() {
	// Create database connection
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})

	// Run migrations
	db.AutoMigrate(&model.Task{})

	// Create simulator
	sim := New().WithTiming(2*time.Second, 3*time.Second)

	// Seed with resources
	seedSimulator(sim)

	// Create Manager
	manager := task.NewManager(db)

	// Replace the k8s client with simulator
	manager.Client = sim

	// Create cluster with simulator
	cluster := &task.Cluster{
		Client: sim,
	}

	// Manager.cluster would normally be initialized from k8s
	// For testing, we inject our simulated cluster
	manager.cluster = cluster

	// Refresh cluster cache (loads from simulator)
	cluster.Refresh()

	// Now the manager will create and monitor pods in the simulator
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Start manager
	go manager.Run(ctx)

	// Create a task in the database
	task := &model.Task{
		Name:  "test-analysis",
		Addon: "analyzer",
		State: "Ready",
		Data: api.Map{
			"mode": api.Map{
				"binary": true,
			},
		},
	}
	db.Create(task)

	// Wait for task to complete
	// The manager will:
	// 1. Pick up the task
	// 2. Create a pod in the simulator
	// 3. Monitor pod state (Pending -> Running -> Succeeded)
	// 4. Update task state accordingly

	time.Sleep(10 * time.Second)

	// Verify task completed
	db.First(task, task.ID)
	// task.State == "Succeeded"
}
*/
