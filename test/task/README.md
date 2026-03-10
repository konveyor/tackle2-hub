# Task Scheduler Test Suite

This directory contains integration tests for the task scheduler using the k8s simulator.

## Test Utilities

The `pkg.go` file provides reusable test utilities:

### Context

The `Context` struct provides a complete test environment:

```go
type Context struct {
    DB       *gorm.DB            // In-memory SQLite database
    Client   client.Client       // K8s simulator client
    Manager  *impTask.Manager    // Task manager instance
    Cancel   context.CancelFunc  // Context cancellation function
    Captured struct {
        TaskFrequency time.Duration  // Saved original frequency
    }
}
```

### Setup and Teardown

```go
func TestMySchedulerFeature(t *testing.T) {
    g := gomega.NewGomegaWithT(t)

    // Setup creates database, k8s client, and configures fast timing
    tc := setup(g)
    defer tc.teardown()

    // Your test code here...
}
```

### Database Seeding

```go
// seed creates common test data:
// - TagCategory "Language"
// - Tag "Java"
// - Application tagged with Java
app := tc.seed(g)
```

### Creating Tasks

```go
// Create tasks manually
task := &model.Task{
    Name:          "test-task",
    Kind:          "analyzer",
    State:         impTask.Ready,
    ApplicationID: &app.ID,
}
err := tc.DB.Create(task).Error
g.Expect(err).To(gomega.BeNil())
```

### Starting the Manager

```go
// newManager creates and starts the task manager in a goroutine
tc.newManager(g)

// The manager will automatically stop when tc.teardown() is called
```

## Example Test

```go
func TestMyFeature(t *testing.T) {
    g := gomega.NewGomegaWithT(t)

    // Setup environment
    tc := setup(g)
    defer tc.teardown()

    // Seed database
    app := tc.seed(g)

    // Create test tasks
    for i := 1; i <= 2; i++ {
        task := &model.Task{
            Name:          "test-task-" + strconv.Itoa(i),
            Kind:          "analyzer",
            State:         impTask.Ready,
            ApplicationID: &app.ID,
        }
        err := tc.DB.Create(task).Error
        g.Expect(err).To(gomega.BeNil())
    }

    // Start manager
    tc.newManager(g)

    // Verify pods were created
    podList := &core.PodList{}
    err := tc.Client.List(context.Background(), podList, &client.ListOptions{
        Namespace: settings.Settings.Hub.Namespace,
    })
    g.Expect(err).To(gomega.BeNil())
    g.Expect(len(podList.Items)).To(gomega.BeNumerically(">=", 1))

    // Wait for completion
    time.Sleep(2 * time.Second)

    // Verify results
    var completedTasks []*model.Task
    err = tc.DB.Find(&completedTasks, "state", impTask.Succeeded).Error
    g.Expect(err).To(gomega.BeNil())
    g.Expect(len(completedTasks)).To(gomega.BeNumerically(">=", 1))
}
```

## Test Configuration

The test environment is configured for fast execution:

- **Task Manager Frequency**: 100ms (default: ~1s)
- **Pod Pending Duration**: 1s
- **Pod Running Duration**: 1s
- **Total test time**: ~2.8s

## K8s Resources

The simulator is pre-seeded with:

- **Addons**: analyzer, language-discovery, platform
- **Extensions**: java, csharp, nodejs, python
- **Tasks**: analyzer, language-discovery, tech-discovery, etc.

See `internal/k8s/seed/resources/` for full resource definitions.

## Running Tests

```bash
# Run all task tests
go test -v ./test/task

# Run specific test
go test -v ./test/task -run TestScheduler

# Run with count for stability testing
go test -v ./test/task -count=3
```
