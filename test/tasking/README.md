# Task Scheduler Test Suite

This directory contains integration tests for the task scheduler using the k8s simulator.

## Test Utilities

The `context.go` file provides reusable test utilities:

### Context

The `Context` struct provides a complete test environment:

```go
type Context struct {
    DB          *gorm.DB           // In-memory SQLite database
    Client      client.Client      // K8s simulator client
    Manager     *task.Manager      // Task manager instance
    Application *model.Application // Pre-seeded test application
    Platform    *model.Platform    // Pre-seeded test platform
}
```

### Creating Test Context

```go
func TestMySchedulerFeature(t *testing.T) {
    g := gomega.NewGomegaWithT(t)

    // New() creates database, k8s client, and seeds test data
    ctx := New(g)

    // Your test code here...
}
```

The `New()` function automatically:
- Creates in-memory SQLite database
- Auto-migrates required tables (Task, TaskGroup, Application, Platform, Bucket, File, TagCategory, Tag)
- Creates k8s simulator with instant pod transitions (0s Pending, 0s Running)
- Seeds test data:
  - TagCategory "Language"
  - Tag "Java"
  - Application "Test Application" (tagged with Java)
  - Platform "Test Platform" (kind: kubernetes)

### Creating Tasks

**Application-based tasks:**

```go
// Create tasks using the pre-seeded application
// Use 'm' for individual task model variables
m := &model.Task{
    Name:          "test-task",
    Kind:          "analyzer",
    State:         task.Ready,
    ApplicationID: &ctx.Application.ID,
}
err := ctx.DB.Create(m).Error
g.Expect(err).To(gomega.BeNil())
```

**Platform-based tasks:**

```go
// Create tasks using the pre-seeded platform
m := &model.Task{
    Name:       "platform-task",
    Kind:       "analyzer",
    State:      task.Ready,
    PlatformID: &ctx.Platform.ID,
}
err := ctx.DB.Create(m).Error
g.Expect(err).To(gomega.BeNil())
```

**Note:** Tasks can reference either an Application OR a Platform as the subject. The RuleUnique and RuleDeps scheduling rules apply to tasks with the same subject (same Application or same Platform).

### Running the Manager

**Synchronous Testing (Recommended)**

Most tests use synchronous reconciliation for deterministic, race-free execution:

```go
// Create manager
ctx.Manager = task.New(ctx.DB, ctx.Client)

// Reconcile once - processes one scheduling cycle
_ = ctx.Manager.Reconcile(context.Background())

// Reconcile until N tasks reach terminal states
ctx.reconcile(g, 2, m1.ID, m2.ID)
```

**Asynchronous Testing (One Test Only)**

Only `TestAsyncManager` uses async mode to verify goroutine lifecycle:

```go
// Configure simulator with realistic timing
ctx.Client = simulator.New().Use(simulator.NewManager(1, 1))

// Start manager in goroutine (async mode)
managerCtx, cancel := context.WithCancel(context.Background())
defer cancel()

go func() {
    ctx.Manager.Run(managerCtx)
}()

time.Sleep(300 * time.Millisecond) // Wait for cluster refresh
```

## Example Tests

### Basic Synchronous Test

```go
func TestMyFeature(t *testing.T) {
    g := gomega.NewGomegaWithT(t)

    // Create test context (database, client, seeded data)
    ctx := New(g)

    // Create test tasks
    var taskIDs []uint
    for i := 1; i <= 2; i++ {
        m := &model.Task{
            Name:          "test-task-" + strconv.Itoa(i),
            Kind:          "analyzer",
            State:         task.Ready,
            ApplicationID: &ctx.Application.ID,
        }
        err := ctx.DB.Create(m).Error
        g.Expect(err).To(gomega.BeNil())
        taskIDs = append(taskIDs, m.ID)
    }

    // Create manager
    ctx.Manager = task.New(ctx.DB, ctx.Client)

    // Reconcile until both tasks complete
    ctx.reconcile(g, 2, taskIDs...)

    // Verify both tasks succeeded
    var tasks []*model.Task
    err := ctx.DB.Find(&tasks, taskIDs).Error
    g.Expect(err).To(gomega.BeNil())
    for _, m := range tasks {
        g.Expect(m.State).To(gomega.Equal(task.Succeeded))
    }
}
```

### Test with Custom Simulator Timing

```go
func TestSlowerTransitions(t *testing.T) {
    g := gomega.NewGomegaWithT(t)

    ctx := New(g)
    // Override with slower transitions: 0s Pending, 2s Running
    ctx.Client = simulator.New().Use(simulator.NewManager(0, 2))

    m := &model.Task{
        Name:          "slow-task",
        Kind:          "analyzer",
        State:         task.Ready,
        ApplicationID: &ctx.Application.ID,
    }
    err := ctx.DB.Create(m).Error
    g.Expect(err).To(gomega.BeNil())

    ctx.Manager = task.New(ctx.DB, ctx.Client)

    // Reconcile twice to reach Running state
    _ = ctx.Manager.Reconcile(context.Background())
    _ = ctx.Manager.Reconcile(context.Background())

    // Verify task is Running (hasn't completed yet)
    err = ctx.DB.First(&m, m.ID).Error
    g.Expect(err).To(gomega.BeNil())
    g.Expect(m.State).To(gomega.Equal(task.Running))
}
```

## Test Configuration

### Synchronous Tests (Default)

Most tests use synchronous testing with instant pod transitions:

- **Pod Pending Duration**: 0s (instant)
- **Pod Running Duration**: 0s (instant)
- **Reconciliation**: Explicit `reconcile()` calls instead of time.Sleep()
- **Total test time**: 0.02-0.10s per test

**Benefits:**
- No race conditions
- Fast execution
- Deterministic behavior
- Easy to debug

### Asynchronous Test (TestAsyncManager)

One test verifies async code paths with realistic timing:

- **Task Manager Frequency**: 100ms
- **Pod Pending Duration**: 1s
- **Pod Running Duration**: 1s
- **Total test time**: ~3.3s

### Simulator Timing Options

```go
// Instant transitions (default, for most tests)
ctx.Client = simulator.New().Use(simulator.NewManager(0, 0))

// Slow transitions (for state-specific tests)
ctx.Client = simulator.New().Use(simulator.NewManager(0, 2)) // 0s Pending, 2s Running
```

### Simulating Pod Failures with TestPodManager

The `TestPodManager` allows testing error scenarios by configuring specific pod failure behaviors:

```go
// Simulate image pull errors (ErrImagePull, ImagePullBackOff, InvalidImageName)
mgr := &TestPodManager{
    imageError: "ErrImagePull",
}
ctx.Client = simulator.New().Use(mgr)

// Simulate container killed (exit code 137) with retry
mgr := &TestPodManager{
    killCount: 1, // Kill once, succeed on retry
}
ctx.Client = simulator.New().Use(mgr)

// Simulate first pod failing immediately
mgr := &TestPodManager{
    failFirstPod: true,
}
ctx.Client = simulator.New().Use(mgr)

// Simulate unschedulable pods (insufficient resources)
mgr := &TestPodManager{
    unschedulable: true,
}
ctx.Client = simulator.New().Use(mgr)
```

**Available configuration fields:**
- `imageError` - Sets image pull error reason (stays Pending, task detects error and fails)
- `killCount` - Number of times to kill pod with exit 137 before allowing success
- `failFirstPod` - Makes first pod transition to Failed immediately
- `unschedulable` - Sets PodReasonUnschedulable condition (capacity exceeded scenario)

**Example test:**
```go
func TestTaskImagePullError(t *testing.T) {
    g := gomega.NewGomegaWithT(t)
    ctx := New(g)

    m := &model.Task{
        Name:          "image-error-task",
        Kind:          "analyzer",
        State:         task.Ready,
        ApplicationID: &ctx.Application.ID,
    }
    err := ctx.DB.Create(m).Error
    g.Expect(err).To(gomega.BeNil())

    // Configure simulator to simulate image pull error
    imageMgr := &TestPodManager{
        imageError: "ErrImagePull",
    }
    ctx.Client = simulator.New().Use(imageMgr)
    ctx.Manager = task.New(ctx.DB, ctx.Client)

    // Task should fail with ImageError event
    ctx.reconcile(g, 1, m.ID)

    var retrieved model.Task
    err = ctx.DB.First(&retrieved, m.ID).Error
    g.Expect(err).To(gomega.BeNil())
    g.Expect(retrieved.State).To(gomega.Equal(task.Failed))
}
```

## Helper Methods

### reconcile(g, n, taskIDs...)

Reconciles until N tasks reach terminal states (Succeeded, Failed, or Canceled):

```go
// Wait for 3 tasks to complete
ctx.reconcile(g, 3, task1.ID, task2.ID, task3.ID)

// Wait for 1 task to complete
ctx.reconcile(g, 1, m.ID)
```

**Parameters:**
- `g`: Gomega instance for assertions
- `n`: Number of tasks expected to reach terminal state
- `taskIDs`: Task IDs to monitor

**Behavior:**
- Calls `Manager.Reconcile()` repeatedly (up to 100 cycles by default)
- Checks if N tasks are terminal after each cycle
- Returns when N tasks complete
- Fails test if max cycles exceeded

## K8s Resources

The simulator is pre-seeded with:

- **Addons**: analyzer, language-discovery, platform
- **Extensions**: java, csharp, nodejs, python
- **Tasks**: analyzer, language-discovery, tech-discovery, etc.

See `internal/k8s/seed/resources/` for full resource definitions.

## Running Tests

```bash
# Run all task tests
go test -v ./test/tasking

# Run specific test
go test -v ./test/tasking -run TestScheduler

# Run with count for stability testing
go test -v ./test/tasking -count=3

# Run with race detector (should pass with synchronous approach)
go test -v -race ./test/tasking
```

## Coding Conventions

### Variable Naming

- **Context**: Always use `ctx` for the test context
- **Task Models**: Use `m` for individual task variables
  ```go
  m := &model.Task{...}
  m1 := &model.Task{...}
  m2 := &model.Task{...}
  ```
- **Specific Tasks**: Use descriptive names
  ```go
  discovery := &model.Task{Kind: "language-discovery", ...}
  analyzer := &model.Task{Kind: "analyzer", ...}
  ```

### Import Style

```go
import (
    "github.com/konveyor/tackle2-hub/internal/task"  // No alias needed
)

// Use task package directly
m.State = task.Ready
ctx.Manager = task.New(ctx.DB, ctx.Client)
```

### Accessing Pre-Seeded Data

```go
// Application and Platform are pre-seeded in context
ApplicationID: &ctx.Application.ID
PlatformID: &ctx.Platform.ID

// Don't create local variables unless needed
```
