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

## Checklist for New Tests

- [ ] Package name is `binding`
- [ ] Imports include: `errors`, `testing`, `shared/api`, `test/cmp`, Gomega
- [ ] Test function named `Test<ResourceName>(t *testing.T)`
- [ ] Gomega initialized: `g := NewGomegaWithT(t)`
- [ ] Dependencies created first with `t.Cleanup()` for cleanup
- [ ] CREATE operation with assertions
- [ ] LIST operation with count validation and `cmp.Eq()` verification
- [ ] First GET (individual) with `cmp.Eq()` verification
- [ ] UPDATE operation modifying multiple fields
- [ ] Second GET with `cmp.Eq()` verification (using `cmp.Eq()`, NOT individual field checks)
- [ ] **ALL resource comparisons use `cmp.Eq()` - NO individual field comparisons**
- [ ] DELETE operation
- [ ] Deletion verified with NotFound error check
- [ ] Section comments for each CRUD+List operation
- [ ] Cleanup uses `t.Cleanup()` and ignores errors with `_ =`
