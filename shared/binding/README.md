# Tackle2 Hub Go Client Library

The `binding` package provides a Go client library for interacting with the Tackle2 Hub REST API. It offers a clean, type-safe interface for managing applications, tasks, assessments, and other Tackle2 resources.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [Working with Resources](#working-with-resources)
- [Advanced Features](#advanced-features)
- [Error Handling](#error-handling)
- [Examples](#examples)
- [API Reference](#api-reference)

## Installation

```bash
go get github.com/konveyor/tackle2-hub/shared/binding
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/konveyor/tackle2-hub/shared/binding"
    "github.com/konveyor/tackle2-hub/shared/api"
)

func main() {
    // Create a new client
    client := binding.New("http://localhost:8080")

    // Authenticate (if required)
    err := client.Login("username", "password")
    if err != nil {
        panic(err)
    }

    // List all applications
    apps, err := client.Application.List()
    if err != nil {
        panic(err)
    }

    for _, app := range apps {
        fmt.Printf("Application: %s (ID: %d)\n", app.Name, app.ID)
    }
}
```

## Authentication

The client supports authentication via username and password:

```go
client := binding.New("http://localhost:8080")

err := client.Login("admin", "password")
if err != nil {
    // Handle authentication error
}
```

Once authenticated, the client automatically includes the authentication token in all subsequent requests.

### Using an Existing Client

You can also provide your own authenticated REST client:

```go
customClient := // ... your RestClient implementation
richClient := binding.New("http://localhost:8080")
richClient.Use(customClient)
```

## Working with Resources

The client provides namespaced APIs for each resource type. All resources follow a consistent CRUD pattern.

### Standard CRUD Operations

Most resources support these standard operations:

```go
// Create
app := &api.Application{Name: "My App"}
err := client.Application.Create(app)

// Get by ID
app, err := client.Application.Get(123)

// List all
apps, err := client.Application.List()

// Update
app.Name = "Updated Name"
err = client.Application.Update(app)

// Delete
err = client.Application.Delete(123)
```

### Available Resources

The client provides access to these resource APIs:

| Resource | Description |
|----------|-------------|
| `Addon` | Manage addons |
| `Analysis` | Application analysis operations |
| `Application` | Application management |
| `Archetype` | Archetype management |
| `Assessment` | Application assessments |
| `Bucket` | Storage bucket operations |
| `BusinessService` | Business service management |
| `ConfigMap` | Configuration maps |
| `Dependency` | Application dependencies |
| `File` | File upload/download |
| `Generator` | Code generation |
| `Identity` | Credential management |
| `Import` | Import operations |
| `JobFunction` | Job functions |
| `Manifest` | Task manifests |
| `MigrationWave` | Migration wave management |
| `Platform` | Platform configuration |
| `Proxy` | Proxy settings |
| `Questionnaire` | Assessment questionnaires |
| `Report` | Report generation |
| `Review` | Application reviews |
| `RuleSet` | Analysis rulesets |
| `Schema` | Schema definitions |
| `Setting` | System settings |
| `Stakeholder` | Stakeholder management |
| `StakeholderGroup` | Stakeholder groups |
| `Tag` | Tag management |
| `TagCategory` | Tag categories |
| `Target` | Analysis targets |
| `Task` | Task management |
| `TaskGroup` | Task grouping |
| `Ticket` | Issue tracking |
| `Tracker` | External tracker integration |

## Advanced Features

### Working with Subresources

Some resources have nested subresources accessed via the `Select()` pattern:

```go
// Select an application
selected := client.Application.Select(appID)

// Access application-specific analysis
analysis, err := selected.Analysis.Get()

// Create an assessment for the application
assessment := &api.Assessment{/* ... */}
err = selected.Assessment.Create(assessment)

// Manage application identities
identities, err := selected.Identity.List()
```

#### Resources with Select() Pattern

- **Application**: Analysis, Assessment, Identity
- **Task**: Bucket, Report, Blocking operations
- **TagCategory**: Tag listing
- **Archetype**: Assessment management
- **TaskGroup**: Bucket operations

### Filtering and Searching

Use filters to narrow down list results:

```go
// Find identities matching a filter
filter := binding.Filter{
    // Add your filter criteria
}
identities, err := client.Identity.Find(filter)
```

### File Operations

#### Uploading Files

```go
// Upload a file
file := &api.File{Name: "report.pdf"}
err := client.File.Create("/path/to/report.pdf", file)
```

#### Downloading Files

```go
// Download to a specific path
err := client.File.Get(fileID, "/path/to/destination.pdf")

// Download to a directory (preserves original filename)
err := client.File.Get(fileID, "/path/to/directory/")
```

### Task Management

Tasks support additional lifecycle operations:

```go
// Create a task
task := &api.Task{/* ... */}
err := client.Task.Create(task)

// Submit a task for execution
err = client.Task.Submit(task.ID)

// Cancel a running task
err = client.Task.Cancel(task.ID)

// Bulk cancel tasks with a filter
err = client.Task.BulkCancel(filter)

// Download task attachments
err = client.Task.GetAttached(task.ID, "/path/to/destination.tar")
```

#### Working with Task Buckets

Tasks have associated storage buckets for files:

```go
selected := client.Task.Select(taskID)

// List bucket contents
entries, err := selected.Bucket.List("/")

// Upload to bucket
err = selected.Bucket.Put("/data/file.txt", "/local/path/file.txt")

// Download from bucket
err = selected.Bucket.Get("/data/file.txt", "/local/destination/")

// Delete from bucket
err = selected.Bucket.Delete("/data/file.txt")
```

#### Task Reports

Create and manage task reports:

```go
selected := client.Task.Select(taskID)

// Create a report
report := &api.TaskReport{/* ... */}
err = selected.Report.Create(report)

// Update report
err = selected.Report.Update(report)

// Delete report
err = selected.Report.Delete()
```

#### Blocking Operations

For operations that need to wait for completion:

```go
import "context"

selected := client.Task.Select(taskID)
ctx := context.Background()

// Cancel and wait for completion
err := selected.Blocking.Cancel(ctx)

// Delete and wait for confirmation
err = selected.Blocking.Delete(ctx)
```

### Settings Management

Settings provide duration-based configuration:

```go
// Get a setting by key
setting, err := client.Setting.GetByKey("task.timeout")

// Update setting
setting.Value = "30m"
err = client.Setting.Update(setting)

// Get duration value
duration, err := client.Setting.Duration("task.timeout")
```

### Identity Management

Identities (credentials) are automatically decrypted:

```go
// List all identities (decrypted)
identities, err := client.Identity.List()

// Get a specific identity (decrypted)
identity, err := client.Identity.Get(identityID)

// Create a new identity
identity := &api.Identity{
    Name: "git-credentials",
    Kind: "git",
    User: "username",
    Password: "password",
}
err = client.Identity.Create(identity)
```

### Tracker Integration

Integrate with external issue trackers:

```go
// List tracker projects
projects, err := client.Tracker.Projects(trackerID)

// Get project issues
issues, err := client.Tracker.ProjectIssues(trackerID, projectID)

// Get a specific issue
issue, err := client.Tracker.Issue(trackerID, issueID)
```

## Error Handling

The binding package provides typed errors for common HTTP status codes:

```go
app, err := client.Application.Get(123)
if err != nil {
    switch {
    case errors.As(err, &binding.NotFound{}):
        // Resource not found (404)
        fmt.Println("Application not found")
    case errors.As(err, &binding.Conflict{}):
        // Conflict error (409)
        fmt.Println("Conflict occurred")
    default:
        // Other error
        fmt.Printf("Error: %v\n", err)
    }
}
```

### Error Types

- `NotFound` - HTTP 404: Resource not found
- `Conflict` - HTTP 409: Resource conflict
- `RestError` - General REST API error
- `EmptyBody` - Response body was empty when content was expected

## Examples

### Example 1: Create and Analyze an Application

```go
package main

import (
    "fmt"
    "github.com/konveyor/tackle2-hub/shared/binding"
    "github.com/konveyor/tackle2-hub/shared/api"
)

func main() {
    client := binding.New("http://localhost:8080")
    client.Login("admin", "password")

    // Create an application
    app := &api.Application{
        Name:        "Customers Portal",
        Description: "Customer management application",
    }
    err := client.Application.Create(app)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Created application ID: %d\n", app.ID)

    // Create an analysis for the application
    selected := client.Application.Select(app.ID)
    analysis := &api.Analysis{
        /* configure analysis */
    }
    err = selected.Analysis.Create(analysis)
    if err != nil {
        panic(err)
    }

    fmt.Println("Analysis created successfully")
}
```

### Example 2: Task Workflow

```go
package main

import (
    "context"
    "fmt"
    "time"
    "github.com/konveyor/tackle2-hub/shared/binding"
    "github.com/konveyor/tackle2-hub/shared/api"
)

func main() {
    client := binding.New("http://localhost:8080")
    client.Login("admin", "password")

    // Create a task
    task := &api.Task{
        Name:  "Analysis Task",
        Addon: "analyzer",
        /* additional configuration */
    }
    err := client.Task.Create(task)
    if err != nil {
        panic(err)
    }

    // Submit the task
    err = client.Task.Submit(task.ID)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Task %d submitted\n", task.ID)

    // Monitor task status
    for {
        task, err = client.Task.Get(task.ID)
        if err != nil {
            panic(err)
        }

        fmt.Printf("Task state: %s\n", task.State)
        if task.State == "Succeeded" || task.State == "Failed" {
            break
        }

        time.Sleep(5 * time.Second)
    }

    // Download task results
    if task.State == "Succeeded" {
        selected := client.Task.Select(task.ID)
        err = selected.Bucket.Get("/output/report.html", "./report.html")
        if err != nil {
            panic(err)
        }
        fmt.Println("Downloaded task report")
    }
}
```

### Example 3: Managing Tags and Categories

```go
package main

import (
    "fmt"
    "github.com/konveyor/tackle2-hub/shared/binding"
    "github.com/konveyor/tackle2-hub/shared/api"
)

func main() {
    client := binding.New("http://localhost:8080")
    client.Login("admin", "password")

    // Create a tag category
    category := &api.TagCategory{
        Name:  "Technology",
        Rank:  1,
    }
    err := client.TagCategory.Create(category)
    if err != nil {
        panic(err)
    }

    // Create tags in the category
    selected := client.TagCategory.Select(category.ID)

    tags := []string{"Java", "Python", "Go"}
    for _, name := range tags {
        tag := &api.Tag{
            Name: name,
            TagCategory: api.Ref{ID: category.ID},
        }
        err = client.Tag.Create(tag)
        if err != nil {
            panic(err)
        }
        fmt.Printf("Created tag: %s\n", name)
    }

    // List all tags in category
    categoryTags, err := selected.Tag.List()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Category '%s' has %d tags\n", category.Name, len(categoryTags))
}
```

### Example 4: Working with Assessments

```go
package main

import (
    "fmt"
    "github.com/konveyor/tackle2-hub/shared/binding"
    "github.com/konveyor/tackle2-hub/shared/api"
)

func main() {
    client := binding.New("http://localhost:8080")
    client.Login("admin", "password")

    // Get or create an application
    app := &api.Application{Name: "My Application"}
    client.Application.Create(app)

    // Create an assessment for the application
    selected := client.Application.Select(app.ID)

    assessment := &api.Assessment{
        Application: &api.Ref{ID: app.ID},
        Questionnaire: api.Ref{ID: 1}, // Reference to questionnaire
        /* additional assessment data */
    }

    err := selected.Assessment.Create(assessment)
    if err != nil {
        panic(err)
    }

    fmt.Printf("Assessment created for application: %s\n", app.Name)

    // List all assessments for the application
    assessments, err := selected.Assessment.List()
    if err != nil {
        panic(err)
    }

    fmt.Printf("Application has %d assessments\n", len(assessments))
}
```

## API Reference

### Core Client

```go
// Create a new client
client := binding.New(baseURL string) *RichClient

// Authenticate
client.Login(user, password string) error

// Use custom REST client
client.Use(client RestClient)
```

### Common Operations

Most resources implement these methods:

```go
Create(r *api.Resource) error
Get(id uint) (*api.Resource, error)
List() ([]api.Resource, error)
Update(r *api.Resource) error
Delete(id uint) error
```

### Filter Operations

Resources that support filtering:

```go
Find(filter Filter) ([]api.Resource, error)
```

### Select Pattern

Resources with subresources:

```go
Select(id uint) Selected
```

## Developer Documentation

For detailed information about the internal architecture, patterns, and conventions used in this package, see [CLAUDE.md](./CLAUDE.md).

## Contributing

When contributing new bindings or modifying existing ones:

1. Follow the patterns documented in [CLAUDE.md](./CLAUDE.md)
2. Maintain consistency with existing code
3. Write comprehensive tests in `test/binding/`
4. Update this README if adding user-facing features

## Support

For issues and questions:
- GitHub Issues: [konveyor/tackle2-hub](https://github.com/konveyor/tackle2-hub/issues)
- Documentation: [Konveyor Documentation](https://konveyor.io/docs/)

## License

See the LICENSE file in the repository root.
