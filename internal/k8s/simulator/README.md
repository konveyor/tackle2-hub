# Kubernetes Cluster Simulator

The simulator provides an in-memory implementation of the Kubernetes client interface for testing the task Manager without requiring an actual Kubernetes cluster.

## Overview

The simulator implements the `sigs.k8s.io/controller-runtime/pkg/client.Client` interface and maintains an in-memory inventory of Kubernetes resources. It's specifically designed to simulate pod lifecycle progression for testing task execution.

## Features

### Supported Resources

The simulator supports the following Kubernetes and custom resources:

- **Core Resources:**
  - `core.Pod` - With automatic state progression
  - `core.Secret`
  - `core.ResourceQuota`

- **Custom Resources (CRDs):**
  - `crd.Tackle`
  - `crd.Addon`
  - `crd.Extension`
  - `crd.Task`

### Pod State Simulation

Pods created in the simulator automatically progress through realistic states over time:

1. **Pending** (default: 10 seconds) - Pod is being scheduled and containers are being created
2. **Running** (default: 20 seconds) - Pod is executing the task
3. **Succeeded** or **Failed** - Pod has completed

The timing and failure behavior are configurable.

## Usage

### Basic Setup

```go
import (
    "github.com/konveyor/tackle2-hub/internal/k8s/simulator"
)

// Create a new simulator with default timing
// Pending: 10s, Running: 20s
// Automatically includes operator-installed resources (Addons, Tasks, Extensions)
sim := simulator.New()

// Inject into Manager or Cluster
manager.Client = sim
```

### Custom Timing

Configure custom durations for pod state transitions:

```go
// Create simulator with fast timing for tests
// Pending: 2s, Running: 3s
sim := simulator.New().WithTiming(2*time.Second, 3*time.Second)
```

### Simulating Failures

Configure the probability of pod failures:

```go
// 20% of pods will fail
sim := simulator.New().WithFailureProbability(0.2)

// All pods will fail
sim := simulator.New().WithFailureProbability(1.0)
```

### Complete Example

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/konveyor/tackle2-hub/internal/k8s/simulator"
    core "k8s.io/api/core/v1"
    meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "sigs.k8s.io/controller-runtime/pkg/client"
)

func main() {
    // Create simulator with custom timing
    sim := simulator.New().WithTiming(5*time.Second, 10*time.Second)

    // Create a pod
    pod := &core.Pod{
        ObjectMeta: meta_v1.ObjectMeta{
            Name:      "my-task-pod",
            Namespace: "konveyor-tackle",
        },
        Spec: core.PodSpec{
            Containers: []core.Container{
                {
                    Name:  "task",
                    Image: "analyzer:latest",
                },
            },
        },
    }

    // Create the pod
    err := sim.Create(context.TODO(), pod)
    if err != nil {
        panic(err)
    }

    // Immediately check status - should be Pending
    retrieved := &core.Pod{}
    sim.Get(context.TODO(), client.ObjectKey{Name: "my-task-pod"}, retrieved)
    fmt.Printf("Phase after creation: %s\n", retrieved.Status.Phase)
    // Output: Phase after creation: Pending

    // Wait 6 seconds - should be Running
    time.Sleep(6 * time.Second)
    sim.Get(context.TODO(), client.ObjectKey{Name: "my-task-pod"}, retrieved)
    fmt.Printf("Phase after 6s: %s\n", retrieved.Status.Phase)
    // Output: Phase after 6s: Running

    // Wait 10 more seconds - should be Succeeded
    time.Sleep(10 * time.Second)
    sim.Get(context.TODO(), client.ObjectKey{Name: "my-task-pod"}, retrieved)
    fmt.Printf("Phase after 16s: %s\n", retrieved.Status.Phase)
    // Output: Phase after 16s: Succeeded
}
```

## Testing with the Simulator

### Unit Tests

```go
func TestTaskManager(t *testing.T) {
    // Create simulator
    sim := simulator.New().WithTiming(1*time.Second, 2*time.Second)

    // Create manager with simulator
    manager := &task.Manager{
        Client: sim,
        // ... other fields
    }

    // Test task execution
    // ...
}
```

### Integration Tests

For integration tests that need to verify task progression:

```go
func TestTaskProgression(t *testing.T) {
    sim := simulator.New().WithTiming(2*time.Second, 3*time.Second)

    // Create task that will use the simulator
    task := createTestTask()

    // Run the task
    started, err := task.Run(&cluster, quota)
    if !started || err != nil {
        t.Fatalf("Task failed to start")
    }

    // Immediately check - task should be Pending
    pod, found := task.Reflect(&cluster)
    if pod.Status.Phase != core.PodPending {
        t.Errorf("Expected Pending, got %s", pod.Status.Phase)
    }

    // Wait for Running state
    time.Sleep(2500 * time.Millisecond)
    pod, found = task.Reflect(&cluster)
    if pod.Status.Phase != core.PodRunning {
        t.Errorf("Expected Running, got %s", pod.Status.Phase)
    }

    // Wait for completion
    time.Sleep(3500 * time.Millisecond)
    pod, found = task.Reflect(&cluster)
    if pod.Status.Phase != core.PodSucceeded {
        t.Errorf("Expected Succeeded, got %s", pod.Status.Phase)
    }
}
```

## Implementation Details

### Architecture

The simulator embeds `k8s.FakeClient` to inherit default no-op implementations of the controller-runtime client interface methods. It then overrides only the methods that need simulation logic:
- `Get()`, `List()`, `Create()`, `Delete()`, `Update()`, `Patch()`

Methods like `Status()`, `Scheme()`, and `RESTMapper()` are inherited from the embedded `FakeClient` without additional code.

### Thread Safety

The simulator uses read-write mutexes to ensure thread-safe access to the in-memory inventory. Multiple goroutines can safely:
- Read resources concurrently
- Create/update/delete resources with proper locking

### Pod State Progression

Pod states are calculated dynamically based on the time elapsed since creation:

- `0 to pendingDuration`: Pod is Pending with ContainerCreating status
- `pendingDuration to (pendingDuration + runningDuration)`: Pod is Running with ready containers
- `After (pendingDuration + runningDuration)`: Pod is Succeeded (or Failed if configured)

The state is calculated on each Get/List operation, so no background goroutines are needed.

**Container Statuses:** The simulator dynamically generates container statuses for all containers defined in `pod.Spec.Containers`. Each container transitions through the same states (Waiting → Running → Terminated) based on the pod's overall timing. This accurately simulates pods with multiple containers, such as those created by the task Manager with main task containers and extension containers.

### Resource Isolation

Each resource type is stored in its own map, keyed by resource name:
- `pods map[string]*podEntry`
- `secrets map[string]*core.Secret`
- `addons map[string]*crd.Addon`
- etc.

### UID Generation

Resources created without a UID are automatically assigned a unique identifier based on creation timestamp.

## Limitations

- Does not simulate network policies or resource quotas enforcement
- Does not validate resource specifications (e.g., container image existence)
- All pods in the same simulator instance use the same timing configuration
- RESTMapper and Scheme methods return nil (inherited from `FakeClient`, sufficient for Manager usage)
- Status subresource operations are inherited from `FakeClient` (no-op behavior)

## Running Tests

```bash
# Run all simulator tests
go test -v ./internal/k8s/simulator

# Run specific test
go test -v ./internal/k8s/simulator -run TestSimulatorPodLifecycle

# Run with race detection
go test -race -v ./internal/k8s/simulator
```

## Integration with Manager

To use the simulator with the task Manager:

```go
import (
    "github.com/konveyor/tackle2-hub/internal/k8s/simulator"
    "github.com/konveyor/tackle2-hub/internal/task"
)

// Create simulator
sim := simulator.New().WithTiming(5*time.Second, 10*time.Second)

// Create Manager with simulator
manager := task.NewManager(db)
manager.Client = sim

// Initialize cluster with simulator
cluster := &task.Cluster{
    Client: sim,
}
cluster.Refresh()  // Will populate from simulator's inventory

// Now Manager will create pods in the simulator
manager.Run(ctx)
```

## Seeding Resources

### Automatic Operator-Installed Resources

When you create a simulator with `New()`, it automatically loads operator-installed resources from YAML files:

```go
sim := simulator.New()
// Automatically includes:
// - Addons (analyzer, language-discovery, platform)
// - Tasks (analyzer, language-discovery, tech-discovery, etc.)
// - Extensions (csharp, java, nodejs, python)
```

These resources are loaded from `data/*.yaml` and represent what the operator installs in a real cluster.

### Adding Cluster-Level Resources

The `seedSimulator()` helper (in `example.go`) adds cluster-level resources that are not operator-installed:

```go
sim := simulator.New()

// Add Tackle CR and ResourceQuota
seedSimulator(sim)
```

The helper adds:
- **Tackle CR**: Main tackle resource
- **ResourceQuota**: Pod quota limits

### Manual Seeding

To manually add resources:

```go
sim := simulator.New()

// Add an addon
addon := &crd.Addon{
    ObjectMeta: meta_v1.ObjectMeta{
        Name:      "analyzer",
        Namespace: "konveyor-tackle",
    },
    Spec: crd.AddonSpec{
        Image: "quay.io/konveyor/analyzer:latest",
    },
}
sim.Create(context.TODO(), addon)

// Add tackle CR
tackle := &crd.Tackle{
    ObjectMeta: meta_v1.ObjectMeta{
        Name:      "tackle",
        Namespace: "konveyor-tackle",
    },
}
sim.Create(context.TODO(), tackle)

// Add a Task CRD
task := &crd.Task{
    ObjectMeta: meta_v1.ObjectMeta{
        Name:      "analyzer",
        Namespace: "konveyor-tackle",
    },
    Spec: crd.TaskSpec{
        Dependencies: []string{"language-discovery"},
        Priority:     10,
    },
}
sim.Create(context.TODO(), task)

// Now these resources will be available via cluster methods
```
