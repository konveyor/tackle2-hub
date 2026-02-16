# Binding Package Patterns and Conventions

This document describes the standard patterns and conventions for writing bindings in `shared/binding/`.

## Package Structure

### Directory Organization

Bindings follow one of two organizational patterns:

#### Pattern 1: Root Package (Simple Resources)
For simple resources without subresources, place the binding directly in `shared/binding/`:
```
shared/binding/
├── file.go          # Simple resource
├── identity.go      # Simple resource
├── setting.go       # Simple resource
└── ...
```

#### Pattern 2: Subpackage (Complex Resources)
For resources with subresources or Select() patterns, create a dedicated package:
```
shared/binding/
├── application/
│   ├── application.go    # Main resource
│   ├── analysis.go       # Subresource
│   ├── assessment.go     # Subresource
│   ├── identity.go       # Subresource
│   └── ...
├── task/
│   └── pkg.go           # All task-related code
├── tagcategory/
│   └── pkg.go           # All tag category code
└── ...
```

**When to use subpackages:**
- Resource has `Select()` pattern with subresources
- Resource has a `Selected` type that would pollute the binding namespace
- Resource has multiple related types that form a cohesive API

## File Naming

### Root Package Files
- One file per resource: `{resource}.go` (e.g., `file.go`, `identity.go`)
- Lowercase, singular form

### Subpackage Files
- **Option 1:** `pkg.go` - All code in one file (preferred for smaller bindings)
- **Option 2:** Multiple files - Split by concern (e.g., `application.go`, `analysis.go`)

## Code Structure

### Standard Binding Pattern

Every binding follows this structure:

```go
package {package}

import (
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding/client"
)

// New creates a new instance (only for subpackages)
func New(c client.RestClient) (h {Resource}) {
	h = {Resource}{client: c}
	return
}

// {Resource} API.
type {Resource} struct {
	client client.RestClient
}

// Create a {Resource}.
func (h {Resource}) Create(r *api.{Resource}) (err error) {
	err = h.client.Post(api.{Resource}sRoute, r)
	return
}

// Get a {Resource} by ID.
func (h {Resource}) Get(id uint) (r *api.{Resource}, err error) {
	r = &api.{Resource}{}
	path := client.Path(api.{Resource}Route).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return
}

// List {Resource}s.
func (h {Resource}) List() (list []api.{Resource}, err error) {
	list = []api.{Resource}{}
	err = h.client.Get(api.{Resource}sRoute, &list)
	return
}

// Update a {Resource}.
func (h {Resource}) Update(r *api.{Resource}) (err error) {
	path := client.Path(api.{Resource}Route).Inject(client.Params{api.ID: r.ID})
	err = h.client.Put(path, r)
	return
}

// Delete a {Resource}.
func (h {Resource}) Delete(id uint) (err error) {
	path := client.Path(api.{Resource}Route).Inject(client.Params{api.ID: id})
	err = h.client.Delete(path)
	return
}
```

### The Select() Pattern

For resources with subresources, implement the Select() pattern:

```go
// Select returns the API for a selected {resource}.
func (h {Resource}) Select(id uint) (h2 Selected) {
	h2 = Selected{
		client:     h.client,
		resourceId: id,
	}
	// Initialize subresources
	h2.Subresource = Subresource{
		client:     h.client,
		resourceId: id,
	}
	return
}

// Selected {resource} API.
type Selected struct {
	client     client.RestClient
	resourceId uint
	Subresource Subresource
}
```

**Examples:**
- `Application.Select(id)` → Access Analysis, Assessment, Identity, etc.
- `Task.Select(id)` → Access Bucket, Report
- `TagCategory.Select(id)` → Access Tag.List()

### Subresource Pattern

Subresources are accessed via Selected:

```go
// Analysis subresource API.
type Analysis struct {
	client client.RestClient
	appId  uint
}

// Create an analysis for the application.
func (h Analysis) Create(r *api.Analysis) (err error) {
	r.Application.ID = h.appId
	path := client.Path(api.AppAnalysesRoute).Inject(client.Params{api.ID: h.appId})
	err = h.client.Post(path, r)
	return
}

// Get the latest analysis for the application.
func (h Analysis) Get() (r *api.Analysis, err error) {
	r = &api.Analysis{}
	path := client.Path(api.AppAnalysisRoute).Inject(client.Params{api.ID: h.appId})
	err = h.client.Get(path, r)
	return
}
```

## Naming Conventions

### Type Names
- Main resource: `{Resource}` (e.g., `Application`, `Task`)
- Selected resource: `Selected` (always)
- Subresource: Descriptive name (e.g., `Analysis`, `Assessment`, `Report`)

### Method Names
- Standard CRUD: `Create`, `Get`, `List`, `Update`, `Delete`
- Bulk operations: Prefix with resource name (e.g., `DeleteList`)
- Specialized operations: Descriptive verb (e.g., `Submit`, `Cancel`, `Archive`)
- Subresource access: `Select(id)`

### Parameter Names
- Resource struct: `r *api.{Resource}`
- ID parameters: `id uint` (for primary), `id1, id2 uint` (for multiple)
- Lists: `list []api.{Resource}`
- Errors: `err error` (always last return)

## Client Usage Patterns

### HTTP Methods

```go
// POST - Create
err = h.client.Post(api.ResourcesRoute, r)

// GET - Retrieve
err = h.client.Get(path, r)

// PUT - Update/Replace
err = h.client.Put(path, r)

// PATCH - Partial Update
err = h.client.Patch(path, r)

// DELETE - Remove
err = h.client.Delete(path)
```

### Path Building

```go
// Simple path
path := client.Path(api.ResourceRoute).Inject(client.Params{api.ID: id})

// Multiple parameters
path := client.Path(api.TrackerProjectRoute).Inject(client.Params{
	api.ID:  trackerId,
	api.ID2: projectId,
})

// Wildcard paths
path := client.Path(api.TaskBucketContentRoute).Inject(client.Params{
	api.ID:       id,
	api.Wildcard: "",
})
```

### File Operations

```go
// Upload file
err = h.client.FilePut(path, source, r)
err = h.client.FilePutEncoded(path, source, r, encoding)

// Download file
err = h.client.FileGet(path, destination)

// Check if destination is directory
isDir, err := h.client.IsDir(destination, false)
if isDir {
	// Construct full path with filename
	destination = filepath.Join(destination, filename)
}
```

## Common Patterns

### Bulk Operations

Bulk operations take a slice of IDs or use filters:

```go
// Bulk delete by IDs
func (h Application) DeleteList(ids []uint) (err error) {
	err = h.client.Delete(api.ApplicationsRoute, ids)
	return
}

// Bulk cancel with filter
func (h Task) BulkCancel(filter client.Filter) (err error) {
	err = h.client.Put(api.TasksCancelRoute, 0, filter.Param())
	return
}
```

### File Downloads

File download methods take a destination parameter:

```go
func (h File) Get(id uint, destination string) (err error) {
	path := client.Path(api.FileRoute).Inject(client.Params{api.ID: id})
	isDir, err := h.client.IsDir(destination, false)
	if err != nil {
		return
	}
	if isDir {
		// Get filename from API
		r := &api.File{}
		err = h.client.Get(path, r)
		if err != nil {
			return
		}
		destination = filepath.Join(destination, r.Name)
	}
	err = h.client.FileGet(path, destination)
	return
}
```

### Report Methods

Report methods are accessed via the Report root resource:

```go
// Global reports
client.Report.Task.Queued()
client.Report.Task.Dashboard()
client.Report.Analysis.RuleReports()

// Per-resource reports via Select()
selected := client.Task.Select(id)
selected.Report.Create(r)
selected.Report.Update(r)
selected.Report.Delete()
```

## Integration with RichClient

### Adding a New Root Binding

1. **Create the binding** (choose root or subpackage)
2. **Import in richclient.go:**
   ```go
   import (
       "github.com/konveyor/tackle2-hub/shared/binding/{resource}"
   )
   ```

3. **Add field to RichClient:**
   ```go
   type RichClient struct {
       {Resource} {package}.{Resource}
   }
   ```

4. **Initialize in build():**
   ```go
   func (r *RichClient) build(client RestClient) {
       r.{Resource} = {package}.New(client)
   }
   ```

5. **Add type alias in pkg.go:**
   ```go
   import "github.com/konveyor/tackle2-hub/shared/binding/{resource}"

   type {Resource} = {package}.{Resource}
   ```

## Documentation Standards

### Type Comments
```go
// {Resource} API.
type {Resource} struct {
	client client.RestClient
}
```

### Method Comments
Use godoc-style comments describing what the method does:

```go
// Create a {Resource}.
// Get a {Resource} by ID.
// List {Resource}s.
// Update a {Resource}.
// Delete a {Resource}.

// Select returns the API for a selected {resource}.
// Upload an analysis manifest at the specified path.
// Archive marks an analysis as archived.
```

## Error Handling

Bindings do not handle errors - they return them directly:

```go
func (h Resource) Get(id uint) (r *api.Resource, err error) {
	r = &api.Resource{}
	path := client.Path(api.ResourceRoute).Inject(client.Params{api.ID: id})
	err = h.client.Get(path, r)
	return  // No error checking, just return
}
```

## Examples

### Simple Resource (Root Package)

See: `shared/binding/file.go`, `shared/binding/identity.go`

### Complex Resource (Subpackage)

See: `shared/binding/application/`, `shared/binding/task/`

### Resource with Select() Pattern

See: `shared/binding/application/application.go`, `shared/binding/tagcategory/pkg.go`

### Resource with Report Subresource

See: `shared/binding/task/pkg.go` (Task.Select().Report)

## Checklist for New Bindings

- [ ] Choose root package vs. subpackage based on complexity
- [ ] Implement standard CRUD methods (Create, Get, List, Update, Delete)
- [ ] Add New() function for subpackages
- [ ] Implement Select() pattern if resource has subresources
- [ ] Use consistent naming (Selected, method names)
- [ ] Follow client usage patterns (path building, file operations)
- [ ] Add to RichClient struct and build() method
- [ ] Add type alias to pkg.go
- [ ] Write comprehensive tests (see test/binding/CLAUDE.md)
- [ ] Document all public types and methods

## Anti-Patterns to Avoid

❌ **Don't:** Create Selected types in root binding package
✓ **Do:** Use subpackages for resources with Selected types

❌ **Don't:** Handle errors in binding methods
✓ **Do:** Return errors directly

❌ **Don't:** Use inconsistent naming (SelectedTask, TaskSelected, etc.)
✓ **Do:** Always use `Selected` for selected resource types

❌ **Don't:** Mix patterns (some subpackages use pkg.go, others use multiple files)
✓ **Do:** Use pkg.go for smaller bindings, split only when necessary

❌ **Don't:** Add business logic to bindings
✓ **Do:** Keep bindings as thin wrappers around HTTP calls

## Migration Guide

When refactoring an existing binding from root to subpackage:

1. Create new subpackage directory
2. Move code to pkg.go in subpackage
3. Add New() function
4. Update imports in richclient.go
5. Update field type in RichClient
6. Update initialization in build()
7. Add type alias in pkg.go
8. Delete old file from root package
9. Verify tests still pass

See commit history for TagCategory refactoring as an example.
