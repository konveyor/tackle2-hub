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

### Method Chaining Patterns

The binding package supports method chaining for certain operations, allowing you to write more concise and fluent code. This is particularly useful when working with subresources that support scoping or filtering.

#### With Method Chaining

Some methods return the modified object, allowing you to chain operations together:

```go
// Fully chained: Select application, set source, and add tag
err := client.Application.Select(appID).
    Tag.Source("analysis").
    Add(tagID)

// Fully chained: Select application, set source, and set fact
err = client.Application.Select(appID).
    Fact.Source("discovery").
    Set("language", "Java")

// Fully chained: Select application, set source, and list tags
tags, err := client.Application.Select(appID).
    Tag.Source("assessment").
    List()

// Fully chained: Select application and get latest analysis
analysis, err := client.Application.Select(appID).
    Analysis.
    Get()

// Fully chained: Select task and upload to bucket
err = client.Task.Select(taskID).
    Bucket.
    Put("/output/report.html", "/local/report.html")
```

#### Without Method Chaining

The same operations can be written without chaining for better readability in complex scenarios:

```go
// Select the application first
selected := client.Application.Select(appID)

// Set up the tag API with source
tagAPI := selected.Tag.Source("analysis")

// Perform multiple operations on the same scope
err := tagAPI.Add(tagID1)
if err != nil {
    return err
}

err = tagAPI.Add(tagID2)
if err != nil {
    return err
}

// List all tags for this source
tags, err := tagAPI.List()
```

#### When to Use Each Approach

**Use chaining when:**
- Performing a single operation on a scoped resource
- The code is simple and readable
- You want concise, one-line operations

**Avoid chaining when:**
- Performing multiple operations on the same scope (reuse the scoped object)
- Error handling needs to be explicit between steps
- The chain becomes too long and hard to read

#### Complete Application Workflow Example

```go
// Complete workflow with mixed chaining approaches
client := binding.New("http://localhost:8080")
client.Login("admin", "password")

// Create application
app := &api.Application{Name: "Customer Portal"}
err := client.Application.Create(app)
if err != nil {
    return err
}

// Use full chaining for single operations
err = client.Application.Select(app.ID).
    Fact.Source("discovery").
    Set("language", "Java")

// Use non-chained for multiple operations on same scope
selected := client.Application.Select(app.ID)
tagAPI := selected.Tag.Source("analysis")
err = tagAPI.Add(languageTagID)
err = tagAPI.Add(frameworkTagID)
err = tagAPI.Add(databaseTagID)

// Upload analysis with full chaining
analysis, err := client.Application.Select(app.ID).
    Analysis.
    Upload("/path/to/manifest.yaml", api.MIMEYAML)

// Download report with full chaining
err = client.Application.Select(app.ID).
    Analysis.
    GetReport("/reports/customer-portal.html")
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

Some resources have nested subresources accessed via the `Select()` pattern. This allows you to work with resources that belong to a specific parent resource.

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

#### Application Subresources

After selecting an application with `client.Application.Select(appID)`, you have access to:

**Analysis** - Manage application analysis results
```go
selected := client.Application.Select(appID)

// Upload analysis manifest (JSON or YAML)
analysis, err := selected.Analysis.Upload("/path/to/manifest.yaml", api.MIMEYAML)

// Create analysis
err = selected.Analysis.Create(&api.Analysis{})

// Get latest analysis
analysis, err = selected.Analysis.Get()

// List all analyses
analyses, err := selected.Analysis.List()

// Download latest analysis report
err = selected.Analysis.GetReport("/path/to/report.html")

// Get insights
insights, err := selected.Analysis.ListInsights()

// Get technology dependencies
deps, err := selected.Analysis.ListDependencies()
```

**Assessment** - Manage application assessments
```go
// Create assessment
err = selected.Assessment.Create(&api.Assessment{})

// List all assessments
assessments, err := selected.Assessment.List()

// Delete assessment
err = selected.Assessment.Delete(assessmentID)
```

**Identity** - Manage credentials for an application
```go
// Create identity
err = selected.Identity.Create(&api.Identity{})

// Get identity
identity, err := selected.Identity.Get(identityID)

// List identities
identities, err := selected.Identity.List()

// Update identity
err = selected.Identity.Update(identity)

// Delete identity
err = selected.Identity.Delete(identityID)
```

**Manifest** - Manage application manifests
```go
// Create manifest
err = selected.Manifest.Create(&api.Manifest{})

// Get latest manifest (with optional decryption/injection)
manifest, err := selected.Manifest.Get(
    client.Param{Key: "decrypted", Value: "1"},
    client.Param{Key: "injected", Value: "1"},
)
```

**Tag** - Associate tags with an application
```go
// Set source for tag operations (optional)
tagAPI := selected.Tag.Source("analysis")

// Add a tag
err = tagAPI.Add(tagID)

// Ensure tag is associated (idempotent)
err = tagAPI.Ensure(tagID)

// List associated tags
tags, err := tagAPI.List()

// Replace all tags for a source
err = tagAPI.Replace([]uint{tagID1, tagID2, tagID3})

// Remove tag association
err = tagAPI.Delete(tagID)
```

**Fact** - Manage application facts (key-value metadata)
```go
// Set source for fact operations
factAPI := selected.Fact.Source("analysis")

// Set a fact (creates or updates)
err = factAPI.Set("language", "Java")

// Get a fact
var language string
err = factAPI.Get("language", &language)

// List all facts for source
facts, err := factAPI.List()

// Create a fact
err = factAPI.Create(&api.Fact{Key: "version", Value: "17"})

// Replace all facts for a source
err = factAPI.Replace(api.Map{"language": "Java", "version": "17"})

// Delete a fact
err = factAPI.Delete("language")
```

**Bucket** - Access application storage bucket
```go
// List bucket contents
entries, err := selected.Bucket.List("/")

// Upload file to bucket
err = selected.Bucket.Put("/data/file.txt", "/local/path/file.txt")

// Download from bucket
err = selected.Bucket.Get("/data/file.txt", "/local/destination/")

// Delete from bucket
err = selected.Bucket.Delete("/data/file.txt")
```

#### Task Subresources

After selecting a task with `client.Task.Select(taskID)`, you have access to:

**Bucket** - Task storage operations
```go
selected := client.Task.Select(taskID)

// List bucket contents
entries, err := selected.Bucket.List("/output")

// Upload to task bucket
err = selected.Bucket.Put("/input/config.yaml", "/local/config.yaml")

// Download from task bucket
err = selected.Bucket.Get("/output/report.html", "./report.html")

// Delete from bucket
err = selected.Bucket.Delete("/temp/data.json")
```

**Report** - Task report management
```go
// Create task report
err = selected.Report.Create(&api.TaskReport{})

// Update task report
err = selected.Report.Update(&api.TaskReport{})

// Delete task report
err = selected.Report.Delete()
```

**Blocking** - Asynchronous task operations
```go
import "context"

ctx := context.Background()

// Cancel task and wait for completion (up to 3 minutes)
err = selected.Blocking.Cancel(ctx)

// Delete task and wait for confirmation (up to 3 minutes)
err = selected.Blocking.Delete(ctx)
```

#### TagCategory Subresources

After selecting a tag category with `client.TagCategory.Select(categoryID)`, you have access to:

**Tag** - List tags in the category
```go
selected := client.TagCategory.Select(categoryID)

// List all tags in this category
tags, err := selected.Tag.List()
```

#### Archetype Subresources

After selecting an archetype with `client.Archetype.Select(archetypeID)`, you have access to:

**Assessment** - Manage archetype assessments
```go
selected := client.Archetype.Select(archetypeID)

// Create assessment for archetype
err = selected.Assessment.Create(&api.Assessment{})

// List assessments
assessments, err := selected.Assessment.List()

// Delete assessment
err = selected.Assessment.Delete(assessmentID)
```

#### TaskGroup Subresources

After selecting a task group with `client.TaskGroup.Select(groupID)`, you have access to:

**Bucket** - Task group storage operations
```go
selected := client.TaskGroup.Select(groupID)

// List bucket contents
entries, err := selected.Bucket.List("/")

// Upload to task group bucket
err = selected.Bucket.Put("/shared/config.yaml", "/local/config.yaml")

// Download from task group bucket
err = selected.Bucket.Get("/shared/data.json", "./data.json")

// Delete from bucket
err = selected.Bucket.Delete("/temp/file.txt")
```

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
