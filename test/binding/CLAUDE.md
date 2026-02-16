# Binding Test Patterns and Conventions

This document describes the standard patterns and conventions for writing binding tests in `test/binding/`.

## File Structure

### Package and Imports
```go
package binding

import (
	"errors"
	"testing"

	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)
```

**Key Points:**
- Package name: `binding`
- Always import `errors` and `testing` from standard library
- Import `github.com/konveyor/tackle2-hub/shared/api` for API types
- Import `github.com/konveyor/tackle2-hub/test/cmp` for comparison utilities
- Use dot import for Gomega: `. "github.com/onsi/gomega"`

## Test Function Pattern

### Naming Convention
- Function name: `Test<ResourceName>(t *testing.T)`
- Example: `TestAnalysisProfile`, `TestIdentity`, `TestStakeholder`
- Uses standard Go test naming with `Test` prefix

### Initialization
```go
func TestResourceName(t *testing.T) {
	g := NewGomegaWithT(t)
	// ... test code
}
```
- Always create Gomega instance: `g := NewGomegaWithT(t)`
- For map[string]any use the api.Map (alias).

## CRUD Test Structure

Tests should follow this standard CRUD+List pattern:

### 1. Setup - Create Dependencies
```go
// Create an identity for the profile to reference
identity := &api.Identity{
	Name: "test-identity",
	Kind: "Test",
}
err := client.Identity.Create(identity)
g.Expect(err).To(BeNil())
t.Cleanup(func() {
	_ = client.Identity.Delete(identity.ID)
})
```

**Key Points:**
- Create dependent resources first
- Add cleanup with `t.Cleanup()` immediately after creation
- Use descriptive comments explaining the dependency relationship
- Ignore cleanup errors: `_ = client.Resource.Delete(id)`

### 2. CREATE - Create the Main Resource
```go
// Define the profile to create
profile := &api.AnalysisProfile{
	Name:        "Test Profile",
	Description: "This is a test analysis profile",
	// ... other fields
}

// CREATE: Create the profile
err = client.AnalysisProfile.Create(profile)
g.Expect(err).To(BeNil())
g.Expect(profile.ID).NotTo(BeZero())

t.Cleanup(func() {
	_ = client.AnalysisProfile.Delete(profile.ID)
})
```

**Key Points:**
- Use section comment: `// CREATE: Create the <resource>`
- Define resource with test data
- Assert no error: `g.Expect(err).To(BeNil())`
- Assert ID is populated: `g.Expect(profile.ID).NotTo(BeZero())`
- Add cleanup with `t.Cleanup()` (main resource cleaned up before dependencies due to LIFO order)

### 3. LIST - List Resources and Verify
```go
// GET: List profiles
list, err := client.AnalysisProfile.List()
g.Expect(err).To(BeNil())
g.Expect(len(list)).To(Equal(1))
eq, report := cmp.Eq(profile, &list[0])
g.Expect(eq).To(BeTrue(), report)
```

**Key Points:**
- Use section comment: `// GET: List <resources>` (note the GET prefix with List action)
- Assert no error
- Assert expected count using `len(list)` with `Equal()` matcher
- **ALWAYS use `cmp.Eq()` to verify the resource** - pass `&list[0]` (pointer to first element)
- This validates that CREATE worked and the resource is retrievable via List

### 4. GET - Retrieve and Verify
```go
// GET: Retrieve the profile and verify it matches
retrieved, err := client.AnalysisProfile.Get(profile.ID)
g.Expect(err).To(BeNil())
g.Expect(retrieved).NotTo(BeNil())
eq, report := cmp.Eq(profile, retrieved)
g.Expect(eq).To(BeTrue(), report)
```

**Key Points:**
- Use section comment: `// GET: Retrieve the <resource> and verify it matches`
- Assert no error
- Assert retrieved object is not nil
- **ALWAYS use `cmp.Eq()` for deep equality comparison** - never compare individual fields
- `cmp.Eq()` returns `(eq bool, report string)`
- Pass report to `Expect()` for better failure messages

### 5. UPDATE - Modify the Resource
```go
// UPDATE: Modify the profile
profile.Name = "Updated Test Profile"
profile.Description = "This is an updated test analysis profile"
// ... modify other fields

err = client.AnalysisProfile.Update(profile)
g.Expect(err).To(BeNil())
```

**Key Points:**
- Use section comment: `// UPDATE: Modify the <resource>`
- Modify multiple fields to test various update scenarios
- Assert no error on update

### 6. GET - Verify Updates
```go
// GET: Retrieve again and verify updates
updated, err := client.AnalysisProfile.Get(profile.ID)
g.Expect(err).To(BeNil())
g.Expect(updated).NotTo(BeNil())
eq, report = cmp.Eq(profile, updated, "UpdateUser")
g.Expect(eq).To(BeTrue(), report)
```

**Key Points:**
- Use section comment: `// GET: Retrieve again and verify updates`
- **ALWAYS use `cmp.Eq()` to verify updates** - never use field-by-field comparisons like `g.Expect(updated.Name).To(Equal(...))`
- May pass additional parameters to `cmp.Eq()` (e.g., `"UpdateUser"`) to exclude certain fields from comparison

### 7. DELETE - Remove the Resource
```go
// DELETE: Remove the profile
err = client.AnalysisProfile.Delete(profile.ID)
g.Expect(err).To(BeNil())

// Verify deletion - Get should fail
_, err = client.AnalysisProfile.Get(profile.ID)
g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
```

**Key Points:**
- Use section comment: `// DELETE: Remove the <resource>`
- Assert no error on delete
- Verify deletion by attempting to Get the resource
- Use `errors.Is(err, &api.NotFound{})` to verify NotFound error

## API Client Pattern

### Global Client Variable
- Tests use a global `client` variable (defined elsewhere in the package)
- Pattern: `client.<ResourceType>.<Method>()`

### Standard Methods
- `Create(resource)` - Creates a resource, populates the ID field
- `List()` - Retrieves all resources of this type, returns a slice
- `Get(id)` - Retrieves a resource by ID
- `Update(resource)` - Updates a resource
- `Delete(id)` - Deletes a resource by ID

## Assertion Patterns

### Error Checking
```go
g.Expect(err).To(BeNil())
```

### Value Validation
```go
g.Expect(profile.ID).NotTo(BeZero())
g.Expect(retrieved).NotTo(BeNil())
```

### List Length Validation
```go
g.Expect(len(list)).To(Equal(1))
g.Expect(len(list)).To(Equal(expectedCount))
```

### Deep Equality - Resource Comparisons

**CRITICAL:** Always use `cmp.Eq()` for comparing resources. Never use individual field comparisons.

```go
eq, report := cmp.Eq(expected, actual)
g.Expect(eq).To(BeTrue(), report)

// With optional ignore fields
eq, report := cmp.Eq(expected, actual, "UpdateUser", "CreateUser")
g.Expect(eq).To(BeTrue(), report)
```

**DO NOT do this:**
```go
// WRONG - Don't compare individual fields
g.Expect(updated.Name).To(Equal(expected.Name))
g.Expect(updated.Description).To(Equal(expected.Description))
```

**DO this instead:**
```go
// CORRECT - Use cmp.Eq() for full resource comparison
eq, report := cmp.Eq(expected, updated, "UpdateUser")
g.Expect(eq).To(BeTrue(), report)
```

### Error Type Checking
```go
g.Expect(errors.Is(err, &api.NotFound{})).To(BeTrue())
```

## Resource Cleanup Pattern

### Cleanup Order
1. Main resource deleted first (in explicit DELETE test section)
2. Dependent resources cleaned up via `t.Cleanup()` functions
3. Cleanup functions execute in LIFO (last-in, first-out) order

### t.Cleanup() Pattern
```go
t.Cleanup(func() {
	_ = client.Resource.Delete(id)
})
```
- Use `t.Cleanup()` with anonymous function for resource cleanup
- Ignore errors with `_` (cleanup is best-effort)
- Preferred over `defer` for Go 1.14+ (more idiomatic for test cleanup)
- Cleanup runs after all subtests complete

## Comments

### Section Comments
Use clear section comments before each CRUD operation:
- `// CREATE: Create the <resource>`
- `// GET: List <resources>` (for List operation)
- `// GET: Retrieve the <resource> and verify it matches` (for individual Get)
- `// UPDATE: Modify the <resource>`
- `// GET: Retrieve again and verify updates`
- `// DELETE: Remove the <resource>`
- `// Verify deletion - Get should fail`

### Dependency Comments
```go
// Create an identity for the profile to reference
```
Explain why dependent resources are needed.

## Test Data Patterns

### Naming
- Use descriptive names: `"Test Profile"`, `"test-identity"`
- For updates, use different values: `"Updated Test Profile"`

### Complex Fields
- Fully populate nested structures to test all fields
- Include both required and optional fields
- Test inclusion and exclusion lists where applicable
- Test references to other resources

## Common Patterns

### Testing References
When a resource references another:
```go
identity := &api.Identity{
	Name: "test-identity",
	Kind: "Test",
}
err := client.Identity.Create(identity)
g.Expect(err).To(BeNil())
t.Cleanup(func() {
	_ = client.Identity.Delete(identity.ID)
})

// Use the reference
profile.Rules.Identity = &api.Ref{
	ID:   identity.ID,
	Name: identity.Name,
}
```

### Testing Lists
```go
Packages: api.InExList{
	Included: []string{"com.example.pkg1", "com.example.pkg2"},
	Excluded: []string{"com.example.pkg3"},
}
```

### Testing Nested Structures
```go
Mode: api.ApMode{
	WithDeps: true,
}
```

## Advanced Patterns

### Subresource Pattern

Many resources have subresources accessed via the `Select()` method:

```go
// Get the selected application API
selected := client.Application.Select(app.ID)

// Create an assessment for the application
err = selected.Assessment.Create(assessment)
g.Expect(err).To(BeNil())

// List tags for the application
tags, err := selected.Tag.List()
g.Expect(err).To(BeNil())

// Access bucket for the application
err = selected.Bucket.Put(tmpFile, "test-file.txt")
g.Expect(err).To(BeNil())
```

**Common Subresources:**
- `Application.Select(id).Assessment` - Application assessments
- `Application.Select(id).Analysis` - Application analyses
- `Application.Select(id).Identity` - Application identities
- `Application.Select(id).Tag` - Application tags
- `Application.Select(id).Fact` - Application facts
- `Application.Select(id).Bucket` - Application bucket/storage
- `Application.Select(id).Manifest` - Application manifests
- `Task.Select(id).Blocking` - Blocking task operations
- `Task.Select(id).Report` - Task reports
- `Bucket.Select(id).Content` - Bucket content operations

### Source-Based Operations

Some subresources support source-based filtering for multi-source data:

```go
// Get source-scoped tag API
source := selected.Tag.Source("T")

// Add tag with source
err = source.Add(1)
g.Expect(err).To(BeNil())

// List tags from specific source
list, err := source.List()
g.Expect(err).To(BeNil())

// Replace all tags for a source
err = source.Replace([]uint{4, 5})
g.Expect(err).To(BeNil())
```

**Resources with Source Support:**
- `Application.Select(id).Tag.Source(name)` - Source-scoped tags
- `Application.Select(id).Fact.Source(name)` - Source-scoped facts

**Key Operations:**
- `Add(id)` - Add a single item
- `Ensure(id)` - Idempotent add
- `Replace([]uint)` - Replace all items for the source
- `List()` - List items from the source
- `Delete(id)` - Delete item (works across sources)

### Ensure Pattern

Some resources support `Ensure()` for idempotent create-or-get operations:

```go
tag := &api.Tag{
	Name: "Test Ensure Tag",
	Category: api.Ref{
		ID:   tagCategory.ID,
		Name: tagCategory.Name,
	},
}

// First call creates the tag
err = client.Tag.Ensure(tag)
g.Expect(err).To(BeNil())
g.Expect(tag.ID).NotTo(BeZero())
firstID := tag.ID

// Second call with same name returns existing tag
tag2 := &api.Tag{
	Name: "Test Ensure Tag",
	Category: api.Ref{
		ID:   tagCategory.ID,
		Name: tagCategory.Name,
	},
}
err = client.Tag.Ensure(tag2)
g.Expect(err).To(BeNil())
g.Expect(tag2.ID).To(Equal(firstID), "Ensure should return existing tag with same name")
```

**Resources with Ensure:**
- `Tag.Ensure(tag)` - Create tag if it doesn't exist, otherwise return existing
- Subresource `Ensure(id)` - Add if not present (idempotent add for associations)

### Blocking Operations with Context

Task operations that may take time use blocking operations with context timeout:

```go
// DELETE with blocking: Remove the task
ctx, cfn := context.WithTimeout(
	context.Background(),
	time.Minute)
defer cfn()
err = client.Task.Select(task.ID).Blocking.Delete(ctx)
g.Expect(err).To(BeNil())

// CANCEL with blocking: Cancel the task
ctx, cfn := context.WithTimeout(
	context.Background(),
	time.Minute)
defer cfn()
err = client.Task.Select(task.ID).Blocking.Cancel(ctx)
g.Expect(err).To(BeNil())
```

**Key Points:**
- Always create a context with timeout
- Use `defer cfn()` to ensure context is canceled
- Use in `t.Cleanup()` for task deletion
- Blocking operations wait for the task to reach final state

**Required Imports:**
```go
import (
	"context"
	"time"
)
```

### Bulk Operations with Filters

Some resources support bulk operations using filters:

```go
import (
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/filter"
)

// Build filter for multiple IDs
f := binding.Filter{}
f.And("id").Eq(filter.Any{task1.ID, task2.ID, task3.ID})

// Bulk cancel tasks matching filter
err = client.Task.BulkCancel(f)
g.Expect(err).To(BeNil())
```

**Key Points:**
- Use `binding.Filter{}` to build filters
- Chain conditions with `.And()` or `.Or()`
- Use `filter.Any{}` for multiple values in equality check
- Bulk operations are typically async

### Bucket/Storage Operations

Bucket content operations use a `.Content` subresource (or direct bucket API):

```go
// Application bucket
selected := client.Application.Select(app.ID)

// Upload a file
err = selected.Bucket.Put(tmpFile, "test-file.txt")
g.Expect(err).To(BeNil())

// Download a file
err = selected.Bucket.Get("test-file.txt", tmpDest)
g.Expect(err).To(BeNil())

// Delete a file
err = selected.Bucket.Delete("test-file.txt")
g.Expect(err).To(BeNil())

// Bucket resource with .Content subresource
bucketSelected := client.Bucket.Select(bucket.ID)
err = bucketSelected.Content.Put(testFile, "uploaded-file.txt")
g.Expect(err).To(BeNil())

// Upload directory
err = bucketSelected.Content.Put(testDir, "uploaded-dir")
g.Expect(err).To(BeNil())
```

**Key Points:**
- Use `Put(sourcePath, destPath)` to upload files or directories
- Use `Get(remotePath, localPath)` to download
- Use `Delete(path)` to remove files or directories
- Supports both file and directory operations

**Required Imports:**
```go
import (
	"os"
	"path/filepath"
)
```

### File Content Verification

Use `assert.EqualFileContent()` to verify file contents match:

```go
import (
	"github.com/konveyor/tackle2-hub/test/assert"
)

// Verify downloaded file matches original
g.Expect(assert.EqualFileContent(downloadedFile, originalFile)).To(BeTrue())
```

### Eventually Pattern for Async Operations

For operations that complete asynchronously, use `g.Eventually()`:

```go
// Wait for tasks to be canceled (async operation)
canceled := []uint{task1.ID, task2.ID, task3.ID}
isDone := func() (done bool) {
	for _, id := range canceled {
		var task *api.Task
		task, err = client.Task.Get(id)
		if err != nil {
			return
		}
		if task.State != tasking.Canceled {
			return
		}
	}
	done = true
	return
}
g.Eventually(isDone, 30*time.Second, time.Second).
	Should(BeTrue(), "Tasks should have been canceled")
```

**Key Points:**
- Define a check function that returns bool
- First param: timeout duration (e.g., `30*time.Second`)
- Second param: polling interval (e.g., `time.Second`)
- Use descriptive message for better failure reporting

### Patch Operations

Some resources support partial updates via `Patch()`:

```go
// Define patch structure
type TaskPatch struct {
	Name string `json:"name"`
}
patch := &TaskPatch{
	Name: "Patched Test Task",
}

// Apply patch
err = client.Task.Patch(task.ID, patch)
g.Expect(err).To(BeNil())

// Update local object to match
task.Name = "Patched Test Task"

// Verify patch
patched, err := client.Task.Get(task.ID)
g.Expect(err).To(BeNil())
eq, report = cmp.Eq(task, patched, "UpdateUser")
g.Expect(eq).To(BeTrue(), report)
```

**Key Points:**
- Define a struct with only fields to update
- Use JSON tags to match API field names
- Update local object to match patch for verification
- Still use `cmp.Eq()` for verification

### Complex Data with api.Map

Use `api.Map` for complex nested data structures:

```go
task := &api.Task{
	Name: "Test Task",
	Kind: "analyzer",
	Data: api.Map{
		"mode": api.Map{
			"binary":       true,
			"withDeps":     false,
			"artifact":     "",
			"diva":         true,
			"csv":          false,
			"dependencies": true,
		},
		"output":  "/windup/report",
		"rules":   []string{"ruleA", "ruleB"},
		"targets": []string{"cloud-readiness"},
		"scope": api.Map{
			"packages": api.Map{
				"included": []string{"com.example"},
				"excluded": []string{"com.example.test"},
			},
		},
	},
}
```

**Key Points:**
- `api.Map` is an alias for `map[string]any`
- Supports arbitrary nesting
- Commonly used for `Data`, `Content`, `Result` fields
- Can mix maps, slices, and primitive values

### Handling Seeded Data in Lists

When testing resources that may have pre-existing seeded data:

```go
// Get existing seeded data before creating test resources
seeded, err := client.Tag.List()
g.Expect(err).To(BeNil())

// Create test resource
tag := &api.Tag{
	Name: "Test Tag",
	Category: api.Ref{
		ID:   tagCategory.ID,
		Name: tagCategory.Name,
	},
}
err = client.Tag.Create(tag)
g.Expect(err).To(BeNil())

// Verify list includes seeded + new resource
list, err := client.Tag.List()
g.Expect(err).To(BeNil())
g.Expect(len(list)).To(Equal(len(seeded) + 1))

// Compare with the newly created item (at end of list)
eq, report := cmp.Eq(tag, list[len(seeded)])
g.Expect(eq).To(BeTrue(), report)
```

### Search API Pattern

Some subresources provide a fluent Search API:

```go
// Application identity search
selected := client.Application.Select(app.ID)

// Search with direct role and indirect kind
search := selected.Identity.Search()
foundIdentity, found, err := search.
	Direct("source").
	Indirect(identity.Kind).
	Find()
g.Expect(err).To(BeNil())
g.Expect(found).To(BeTrue())

// Chain multiple Direct calls (OR logic)
foundIdentity, found, err = search.
	Direct("role1").
	Direct("role2").
	Indirect(kind).
	Find()
```

**Key Points:**
- `Search()` returns a builder object
- Chain methods like `Direct(role)` and `Indirect(kind)`
- Multiple `Direct()` calls act as OR
- Call `Find()` to execute the search
- Returns `(resource, found bool, error)`

### Encrypted Data Handling

For resources with encrypted secrets (e.g., Manifest):

```go
// Create manifest with secrets
manifest := &api.Manifest{
	Application: api.Ref{ID: app.ID},
	Content: api.Map{
		"key":      "$(key)",
		"password": "$(password)",
	},
	Secret: api.Map{
		"key":      "ABCDEF",
		"password": "1234",
	},
}
err = selected.Manifest.Create(manifest)
g.Expect(err).To(BeNil())

// After create, Secret is encrypted (doesn't match original)

// GET with default returns encrypted secret
encrypted, err := selected.Manifest.Get()
g.Expect(err).To(BeNil())

// GET with Decrypted param returns decrypted secret
import "github.com/konveyor/tackle2-hub/shared/binding"

decrypted, err := selected.Manifest.Get(
	binding.Param{Key: api.Decrypted, Value: "1"})
g.Expect(err).To(BeNil())

// GET with Decrypted and Injected returns content with secrets injected
injected, err := selected.Manifest.Get(
	binding.Param{Key: api.Decrypted, Value: "1"},
	binding.Param{Key: api.Injected, Value: "1"})
g.Expect(err).To(BeNil())
// injected.Content has $(key) replaced with "ABCDEF", etc.
```

**Key Points:**
- Secrets are automatically encrypted on create/update
- Use `binding.Param{Key: api.Decrypted, Value: "1"}` to retrieve decrypted
- Use `binding.Param{Key: api.Injected, Value: "1"}` to inject secrets into content
- Content uses `$(variableName)` syntax for secret placeholders

### Attachment Handling

Tasks can have attached files that can be downloaded as a tarball:

```go
// Create files
file1, err := client.File.Put(tmpFile1)
g.Expect(err).To(BeNil())

file2, err := client.File.Put(tmpFile2)
g.Expect(err).To(BeNil())

// Create task with attachments
task := &api.Task{
	Name: "Test Task",
	Attached: []api.Attachment{
		{
			ID:   file1.ID,
			Name: "file1.txt",
		},
		{
			ID:   file2.ID,
			Name: "file2.txt",
		},
	},
}
err = client.Task.Create(task)
g.Expect(err).To(BeNil())

// Download attachments as tarball
err = client.Task.GetAttached(task.ID, "/tmp/attachments.tar.gz")
g.Expect(err).To(BeNil())

// Verify tarball was created
info, err := os.Stat("/tmp/attachments.tar.gz")
g.Expect(err).To(BeNil())
g.Expect(info.Size()).To(BeNumerically(">", 0))
```

### Additional Imports for Advanced Tests

Depending on test requirements, you may need additional imports:

```go
package binding

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	tasking "github.com/konveyor/tackle2-hub/internal/task"
	"github.com/konveyor/tackle2-hub/shared/api"
	"github.com/konveyor/tackle2-hub/shared/binding"
	"github.com/konveyor/tackle2-hub/shared/binding/filter"
	"github.com/konveyor/tackle2-hub/test/assert"
	"github.com/konveyor/tackle2-hub/test/cmp"
	. "github.com/onsi/gomega"
)
```

**Common Additional Imports:**
- `"context"` - For blocking operations
- `"time"` - For timeouts and Eventually polling
- `"os"` - For file operations
- `"path/filepath"` - For path manipulation
- `tasking "github.com/konveyor/tackle2-hub/internal/task"` - For task state constants
- `"github.com/konveyor/tackle2-hub/shared/binding"` - For Param and Filter
- `"github.com/konveyor/tackle2-hub/shared/binding/filter"` - For filter.Any
- `"github.com/konveyor/tackle2-hub/test/assert"` - For file content comparison

## Common Anti-Patterns to Avoid

### ❌ Individual Field Comparisons After Update

**DON'T:**
```go
// After update, verify fields one-by-one
updated, err := client.Resource.Get(resource.ID)
g.Expect(err).To(BeNil())
g.Expect(updated.Field1).To(Equal(resource.Field1))
g.Expect(updated.Field2).To(Equal(resource.Field2))
// ... more field comparisons
```

**DO:**
```go
// After update, use cmp.Eq() for full comparison
updated, err := client.Resource.Get(resource.ID)
g.Expect(err).To(BeNil())
eq, report := cmp.Eq(resource, updated, "UpdateUser")
g.Expect(eq).To(BeTrue(), report)
```

### ❌ Not Handling Seeded Data

**DON'T:**
```go
list, err := client.Tag.List()
g.Expect(err).To(BeNil())
g.Expect(len(list)).To(Equal(1))  // Fails if seeded data exists
```

**DO:**
```go
seeded, err := client.Tag.List()
g.Expect(err).To(BeNil())

// ... create test tag ...

list, err := client.Tag.List()
g.Expect(err).To(BeNil())
g.Expect(len(list)).To(Equal(len(seeded) + 1))
```

### ❌ Forgetting Context Timeout for Blocking Operations

**DON'T:**
```go
err = client.Task.Select(task.ID).Blocking.Delete(context.Background())
// Potential indefinite hang
```

**DO:**
```go
ctx, cfn := context.WithTimeout(context.Background(), time.Minute)
defer cfn()
err = client.Task.Select(task.ID).Blocking.Delete(ctx)
```

## Checklist for New Tests

- [ ] Package name is `binding`
- [ ] Imports include: `errors`, `testing`, `shared/api`, `test/cmp`, Gomega
- [ ] Additional imports as needed: `context`, `time`, `os`, `binding`, etc.
- [ ] Test function named `Test<ResourceName>(t *testing.T)`
- [ ] Gomega initialized: `g := NewGomegaWithT(t)`
- [ ] Dependencies created first with `t.Cleanup()` for cleanup
- [ ] CREATE operation with assertions
- [ ] LIST operation with count validation and `cmp.Eq()` verification
- [ ] Handle seeded data if applicable (get baseline before creating test data)
- [ ] First GET (individual) with `cmp.Eq()` verification
- [ ] UPDATE operation modifying multiple fields
- [ ] Second GET with `cmp.Eq()` verification (using `cmp.Eq()`, NOT individual field checks)
- [ ] **ALL resource comparisons use `cmp.Eq()` - NO individual field comparisons**
- [ ] DELETE operation (with context timeout for tasks)
- [ ] Deletion verified with NotFound error check
- [ ] Section comments for each CRUD+List operation
- [ ] Cleanup uses `t.Cleanup()` and ignores errors with `_ =`
- [ ] For tasks: use `Blocking.Delete(ctx)` with timeout in cleanup
- [ ] For subresources: use `Select()` pattern
- [ ] For source-based operations: use `Source()` pattern
- [ ] For async operations: use `Eventually()` with timeout
