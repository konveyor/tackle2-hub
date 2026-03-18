# Addon Adapter

The Addon Adapter provides a Go framework for building Tackle2 Hub addons. It simplifies addon development by providing:

- **Task lifecycle management** - Automatic reporting of task status (Started, Succeeded, Failed)
- **Hub API integration** - Pre-configured clients for all hub resources
- **Progress tracking** - Built-in support for activity logs, progress updates, and file attachments
- **Error handling** - Panic recovery and automatic error reporting
- **Resource injection** - Identity and proxy injection for SCM operations

## Quick Start

```go
package main

import (
    hub "github.com/konveyor/tackle2-hub/shared/addon"
)

var (
    // Global addon adapter instance
    addon = hub.Addon
)

func main() {
    addon.Run(func() (err error) {
        // Get the application associated with this task
        application, err := addon.Task.Application()
        if err != nil {
            return
        }

        // Report activity
        addon.Activity("Processing application: %s", application.Name)

        // Do work...

        return
    })
}
```

## Core Concepts

### The `addon.Run()` Pattern

`addon.Run()` is the main entry point for all addons. It:
1. Loads the task from the hub
2. Reports the task as **Started**
3. Executes your addon function
4. Reports **Succeeded** or **Failed** based on the result
5. Handles panics and converts them to task failures

```go
addon.Run(func() (err error) {
    // Your addon logic here
    return
})
```

**Note:** `addon.Run()` calls `os.Exit(1)` on failure. Your addon function should return errors rather than calling `os.Exit()` directly.

### Task Data

Addons receive structured input data through the task:

```go
type MyData struct {
    Path    string `json:"path"`
    Options []string `json:"options"`
}

d := &MyData{}
err := addon.DataWith(d)
if err != nil {
    return
}

// Use the data
addon.Activity("Processing path: %s", d.Path)
```

## Progress Reporting

### Activity Logs

Report what the addon is currently doing using **printf-style formatting**:

```go
addon.Activity("Analyzing source code...")
addon.Activity("Found %d files to process", fileCount)
addon.Activity("Processing file: %s (size: %d bytes)", filename, size)
```

Multi-line activity entries are automatically formatted:
```go
addon.Activity("Line 1\nLine 2\nLine 3")
// Reports as:
// Line 1
// > Line 2
// > Line 3
```

### Progress Tracking

Report total work and completion:

```go
files := []string{"file1", "file2", "file3"}

// Set the total number of items
addon.Total(len(files))

for _, file := range files {
    processFile(file)

    // Increment completed count
    addon.Increment()
}

// Or set completed directly
addon.Completed(10)
```

### Error Reporting

Report non-fatal errors while continuing execution:

```go
// Report with severity
addon.Error(api.TaskError{
    Severity:    "Warning",
    Description: "Could not process optional configuration",
})

// Or use printf-style formatting
addon.Errorf("Error", "Failed to analyze file %s: %v", filename, err)
addon.Errorf("Warning", "Skipping %d files due to permissions", skipCount)
```

### File Attachments

Attach files to activity entries:

```go
// Upload a file to the hub
file, err := addon.File.Put("/tmp/report.html")
if err != nil {
    return
}

// Attach to the last activity entry
addon.Attach(file)

// Or attach to a specific activity (1-based index)
addon.AttachAt(file, 3)
```

### Task Results

Store structured results with the task:

```go
addon.Result(api.Map{
    "filesProcessed": 42,
    "warnings": 3,
    "duration": "5m30s",
})
```

## Working with Hub Resources

### Application

```go
// Get the task's application
application, err := addon.Task.Application()
if err != nil {
    return
}

// Access application properties
addon.Activity("Processing: %s", application.Name)

// Update application facts
facts := addon.Application.Facts(application.ID)
facts.Source("my-addon")
err = facts.Set("analyzed", true)

// Work with application tags
tags := addon.Application.Tags(application.ID)
tags.Source("my-addon")
err = tags.Ensure(tagID)

// Access application bucket
bucket := addon.Application.Select(application.ID).Bucket
err = bucket.Put("/tmp/report", "reports/analysis.html")
```

### Platform

```go
// Get the task's platform
platform, err := addon.Task.Platform()
if err != nil {
    return
}

addon.Activity("Running on platform: %s", platform.Name)
```

### Addon and Extensions

```go
// Get addon metadata with environment injection
addonMeta, err := addon.Task.Addon(true)
if err != nil {
    return
}

// Access extension metadata (injected with environment variables)
for _, ext := range addonMeta.Extensions {
    metadata := ext.Metadata.(map[string]any)
    // Use injected configuration
}
```

### Tags

```go
// Ensure tag category exists
category := &api.TagCategory{
    Name:  "Analysis",
    Color: "#2b9af3",
}
err = addon.TagCategory.Ensure(category)
if err != nil {
    return
}

// Ensure tag exists
tag := &api.Tag{
    Name: "Analyzed",
    Category: api.Ref{ID: category.ID},
}
err = addon.Tag.Ensure(tag)
if err != nil {
    return
}

// Associate with application
tags := addon.Application.Tags(application.ID)
tags.Source("my-addon")
err = tags.Ensure(tag.ID)
```

### Files

```go
// Upload a file
file, err := addon.File.Put("/etc/hosts")
if err != nil {
    return
}

// Download a file
err = addon.File.Get(file.ID, "/tmp/downloaded")
if err != nil {
    return
}

// Download to directory (filename preserved)
err = addon.File.Get(file.ID, "/tmp")
if err != nil {
    return
}

// Delete a file
err = addon.File.Delete(file.ID)
```

### Bucket

```go
// Get application bucket
bucket := addon.Application.Select(application.ID).Bucket

// Upload file or directory
err = bucket.Put("/tmp/report.html", "reports/analysis.html")
err = bucket.Put("/tmp/output-dir", "reports/detailed")

// Download file or directory
err = bucket.Get("reports/analysis.html", "/tmp/downloaded.html")
err = bucket.Get("reports/detailed", "/tmp/output-dir")

// Delete file or directory
err = bucket.Delete("reports/analysis.html")
```

## Source-Based Resources

Some resources support multiple sources (e.g., tags and facts from different addons):

### Tags with Source

```go
tags := addon.Application.Tags(application.ID)
tags.Source("analyzer")

// Add tags from this source
err = tags.Ensure(tag1.ID)
err = tags.Ensure(tag2.ID)

// Replace ALL tags from this source
err = tags.Replace([]uint{tag1.ID, tag2.ID})
```

### Facts with Source

```go
facts := addon.Application.Facts(application.ID)
facts.Source("analyzer")

// Set individual facts
err = facts.Set("language", "Java")
err = facts.Set("frameworks", []string{"Spring", "Hibernate"})

// Get a fact
var language string
err = facts.Get("language", &language)

// Replace ALL facts from this source
err = facts.Replace(api.Map{
    "language": "Java",
    "version": "11",
    "frameworks": []string{"Spring", "Hibernate"},
})
```

## SCM Integration

Use the `shared/addon/scm` package for Git/Subversion repository operations with automatic identity and proxy injection:

```go
import (
    hub "github.com/konveyor/tackle2-hub/shared/addon"
    "github.com/konveyor/tackle2-hub/shared/addon/scm"
)

// Get application with repository
application, err := addon.Task.Application()
if err != nil {
    return
}

// Get identity for the repository (optional)
identity, found, err := addon.Application.Select(application.ID).
    Identity.Search().
    Direct("source").
    Indirect("git").
    Find()
if err != nil {
    return
}
if !found {
    identity = nil
}

// Create repository instance
repo, err := scm.New("/tmp/source", application.Repository, identity)
if err != nil {
    return
}

// Clone repository
err = repo.Fetch()
if err != nil {
    return
}
```

The SCM package automatically:
- Injects identity credentials (user/password/key)
- Configures proxies from hub settings
- Sets up insecure mode based on hub settings (`git.insecure.enabled`, `svn.insecure.enabled`)
- Supports both Git and Subversion repositories

**Note:** See the [scm package documentation](../../scm/) for advanced repository operations.

## Command Execution

Use the `shared/addon/command` package for executing external commands with automatic **task activity reporting**, output capture, and attachment:

```go
import (
    hub "github.com/konveyor/tackle2-hub/shared/addon"
    "github.com/konveyor/tackle2-hub/shared/addon/command"
)

// Create command
cmd := command.New("mvn")
cmd.Options.Add("clean")
cmd.Options.Add("install")
cmd.Dir = "/tmp/source"

// Execute command
err := cmd.Run()
if err != nil {
    return
}
```

The command package automatically:
- **Reports command execution to task activity** (started, succeeded, failed)
- Captures stdout/stderr to a file
- Attaches the output file to the task activity
- Handles command errors

**Building command options:**

```go
// Progressive construction with Add()
cmd.Options.Add("--settings")
cmd.Options.Add("/path/to/settings.xml")

// Formatted options with Addf()
cmd.Options.Addf("-Dmaven.repo.local=%s", repoPath)
cmd.Options.Addf("-DskipTests=%t", skipTests)

// Or assign directly
cmd.Options = []string{"clean", "install", "-DskipTests"}
```

**Command fields:**
- `Options` - Command arguments with `Add()` and `Addf()` methods
- `Dir` - Working directory
- `Env` - Environment variables ([]string)

## Error Handling

### Return Errors, Don't Exit

```go
// ✅ Good - return errors
addon.Run(func() (err error) {
    result, err := doSomething()
    if err != nil {
        return err  // Let addon.Run() handle it
    }
    return
})

// ❌ Bad - don't exit directly
addon.Run(func() (err error) {
    result, err := doSomething()
    if err != nil {
        os.Exit(1)  // Don't do this!
    }
    return
})
```

### Panic Recovery

`addon.Run()` automatically recovers from panics:

```go
addon.Run(func() (err error) {
    // If this panics, it's caught and reported as a failure
    riskyOperation()
    return
})
```

### Explicit Failure

You can explicitly fail the task while continuing work using **printf-style formatting**:

```go
addon.Run(func() (err error) {
    result, err := criticalOperation()
    if err != nil {
        // Mark as failed but continue cleanup (printf-style)
        addon.Failed("Critical operation failed: %v", err)
    }

    // Cleanup work here
    cleanup()

    return
})
```

## Advanced Patterns

### Custom Status Reporting

If you explicitly set task status, `addon.Run()` won't override it:

```go
addon.Run(func() (err error) {
    // Do some work...

    if partialSuccess {
        // Explicitly mark as succeeded with warnings
        addon.Error(api.TaskError{
            Severity: "Warning",
            Description: "Partial success",
        })
        addon.Succeeded()
        return  // Run() won't call Succeeded() again
    }

    return  // Run() calls Succeeded() automatically
})
```

### Working with Identities

```go
// Get application
application, err := addon.Task.Application()
if err != nil {
    return
}

// Get identity search API
search := addon.Application.Select(application.ID).Identity.Search()

// Search for direct identity (assigned to application)
identity, found, err := search.Direct("source").Find()

// Or search for indirect identity (default for kind)
identity, found, err := search.Indirect("git").Find()

// Or chain multiple searches
identity, found, err := search.
    Direct("source").
    Direct("credentials").
    Indirect("maven").
    Find()
```

### Environment Injection

Extension metadata can reference environment variables that are automatically injected:

```json
{
  "metadata": {
    "database": {
      "host": "$(DB_HOST)",
      "port": "$(DB_PORT)"
    }
  }
}
```

When the addon loads with injection enabled, these are replaced with actual values from the environment.

## API Reference

### Global Variables

- `addon` - The global adapter instance (`hub.Addon`)
- `addon.Log` - Structured logger (`logr.Logger`)
- `addon.Wrap` - Error wrapper with stack traces

### Task API

| Method | Description |
|--------|-------------|
| `Load()` | Load task from hub (called automatically by `Run()`) |
| `Application()` | Get the task's application |
| `Platform()` | Get the task's platform |
| `Addon(inject bool)` | Get addon metadata with optional injection |
| `Data()` | Get raw task data as `any` |
| `DataWith(object)` | Unmarshal task data into struct |
| `Bucket()` | Get task bucket API |

### Reporting API

| Method | Description |
|--------|-------------|
| `Started()` | Report task started (called by `Run()`) |
| `Succeeded()` | Report task succeeded (called by `Run()`) |
| `Failed(reason, ...)` | Report task failed with reason (**printf-style**) |
| `Activity(entry, ...)` | Report activity (**printf-style**) |
| `Error(...TaskError)` | Report errors |
| `Errorf(severity, description, ...)` | Report formatted error (**printf-style**) |
| `Total(n)` | Set total items to process |
| `Increment()` | Increment completed count |
| `Completed(n)` | Set completed count |
| `Attach(file)` | Attach file to last activity |
| `AttachAt(file, activity)` | Attach file to specific activity |
| `Result(object)` | Store structured result |

### Resource APIs

All hub resource APIs are available through `addon.*`:
- `Application` - Application CRUD and subresources
- `AnalysisProfile` - Analysis profiles
- `Archetype` - Archetypes
- `File` - File upload/download
- `Generator` - Generators
- `Identity` - Identity management
- `Manifest` - Application manifests
- `Platform` - Platform definitions
- `Proxy` - Proxy configuration
- `RuleSet` - Analysis rulesets
- `Schema` - Schemas
- `Setting` - Hub settings
- `Tag` - Tags
- `TagCategory` - Tag categories
- `Target` - Analysis targets

See the [binding documentation](../../binding/) for detailed API reference.

## Examples

Complete addon examples can be found in:
- [`hack/cmd/addon/main.go`](../../../hack/cmd/addon/main.go) - Example addon demonstrating most features
- [Analyzer addon](https://github.com/konveyor/tackle2-addon-analyzer) - Production analyzer addon

## Testing

See [`pkg_test.go`](./pkg_test.go) for comprehensive unit test examples using the stub pattern to mock hub responses.

## Best Practices

1. **Always use `addon.Run()`** - Don't manage task lifecycle manually
2. **Return errors** - Let `Run()` handle error reporting and exit
3. **Report activity frequently** - Help users understand addon progress
4. **Use source-based operations** - When working with tags/facts from multiple sources
5. **Handle missing resources gracefully** - Not all tasks have applications/platforms
6. **Use structured logging** - `addon.Log.Info()`, `addon.Log.Error()`
7. **Attach important files** - Make reports visible in the UI
8. **Set meaningful task results** - Store structured output for later use
