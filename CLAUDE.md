# Tackle2 Hub Coding Patterns Conventions and Standards

This document defines the Go coding standards and patterns for the Tackle2 Hub project. Follow these conventions to maintain consistency and code quality across the codebase.

## Table of Contents

- [Project Organization](#project-organization)
- [Package Organization](#package-organization)
- [Object-Oriented Design](#object-oriented-design)
- [Function Design](#function-design)
- [Function and Method Standards](#function-and-method-standards)
  - [Naming Conventions](#naming-conventions)
  - [Function Returns](#function-returns)
  - [Code Documentation and Comments](#code-documentation-and-comments)
- [Testing](#testing)
- [Error Handling](#error-handling)
- [Git Commits](#git-commits)

---

## Project Organization

### Quick Decision Guide

When creating a new type, ask:

1. **Will external clients/addons need this?** → `shared/`
2. **Is this a REST API request/response type?** → `shared/api/`
3. **Is this a database model?** → `internal/model/`
4. **Is this handler logic?** → `internal/api/`
5. **Does it import anything from `internal/`?** → Must stay in `internal/`

### Exported Types Go in shared/

**All exported types, interfaces, and APIs that are consumed by external packages or clients must live in `shared/`.** This includes:

- REST API request/response types
- Client binding interfaces and implementations
- Public configuration structures
- Addon framework types and interfaces
- Any type that external code needs to import

```go
// Good - exported types in shared/
// File: shared/api/application.go
package api

// Application REST resource - exported for client use
type Application struct {
    Resource        `yaml:",inline"`
    Name            string        `json:"name" binding:"required"`
    Description     string        `json:"description"`
    BusinessService *Ref          `json:"businessService"`
    Owner           *Ref          `json:"owner"`
    Tags            []TagRef      `json:"tags"`
    // ... other exported fields
}

// File: shared/binding/application.go
package binding

// Application client - exported for external consumers
type Application struct {
    client *Client
}

func (h *Application) Create(r *api.Application) (err error) {
    // Client implementation
    return
}
```

```go
// Bad - putting exported client types in internal/
// File: internal/api/client.go
package api

// Don't export client-facing types from internal/
type ApplicationClient struct {
    // ...
}
```

### Internal Implementation Details

**Implementation details, handler logic, and hub-specific code stays in `internal/`.** This includes:

- HTTP handler implementations
- Database models and migrations
- Business logic and domain services
- Internal utilities and helpers

```go
// Good - internal handler implementation
// File: internal/api/application.go
package api

import (
    "github.com/konveyor/tackle2-hub/internal/model"
    "github.com/konveyor/tackle2-hub/shared/api"  // Import from shared OK
)

// ApplicationHandler - internal HTTP handler
type ApplicationHandler struct {
    BaseHandler
}

func (h ApplicationHandler) Get(ctx *gin.Context) {
    m := &model.Application{}  // Internal model
    // ... handler implementation
    r := api.Application{}     // Shared API type for response
    h.Respond(ctx, http.StatusOK, r)
}
```

### Critical Rule: shared/ Cannot Import from internal/

**NEVER import anything from `internal/` in `shared/` packages.** This is a hard architectural boundary that ensures:
- Clean separation of concerns
- Clients don't accidentally depend on internal implementation details
- Internal refactoring doesn't break external consumers

```go
// FORBIDDEN - shared/ importing from internal/
// File: shared/api/application.go
package api

import (
    "github.com/konveyor/tackle2-hub/internal/model"  // ❌ NEVER DO THIS
)

type Application struct {
    Model model.Application  // ❌ Exposing internal types
}
```

```go
// Correct - shared/ is independent
// File: shared/api/application.go
package api

type Application struct {
    Name            string   `json:"name"`
    BusinessService *Ref     `json:"businessService"`
    // Only reference other shared/ types
}
```

### Import Direction

The allowed import direction is:

```
internal/  →  shared/     ✅ OK: Internal can import shared types
shared/    →  internal/   ❌ FORBIDDEN: Shared cannot import internal
external   →  shared/     ✅ OK: External clients import shared types
external   →  internal/   ❌ FORBIDDEN: Enforced by Go visibility rules
```

### Example: Correct Usage

```go
// ✅ Internal handler uses shared API types
// File: internal/api/application.go
package api

import (
    "github.com/konveyor/tackle2-hub/internal/model"
    "github.com/konveyor/tackle2-hub/shared/api"
)

func (h ApplicationHandler) Create(ctx *gin.Context) {
    r := &api.Application{}  // Shared API type
    err := h.Bind(ctx, r)
    m := r.Model()           // Convert to internal model
    // ... create in database
}
```

```go
// ✅ Shared binding has no internal dependencies
// File: shared/binding/application.go
package binding

import "github.com/konveyor/tackle2-hub/shared/api"

func (h *Application) Create(r *api.Application) (err error) {
    // Makes HTTP call, no knowledge of internal implementation
    return
}
```

---

## Package Organization

### Meaningful Package Names

**Use descriptive, domain-specific package names** that clearly indicate their purpose. Package names should be:
- **Singular nouns** (e.g., `model`, `api`, `filter`)
- **Descriptive** of their domain responsibility
- **Short** and concise

```go
// Good package names
package api          // REST API handlers
package model        // Data models
package assessment   // Assessment logic
package filter       // Query filtering
package resource     // REST resources
package task         // Task management
package tracker      // Issue tracker integration
package trigger      // Event triggers
```

### Prohibited Package Names

**Never use these generic package names:**
- `helper` - Too vague; move functions to domain-specific packages
- `common` - Too broad; use specific names like `shared` with subdirectories
- `util` or `utils` - Too generic; use domain-specific names
- `misc` - Indicates poor organization

```go
// Bad
package helper
package common
package util
package utils
package misc

// Good - use specific names instead
package validation
package conversion
package formatting
package authentication
```

### Package Structure

Organize packages by domain responsibility:

```
internal/
├── api/           # REST API handlers
│   ├── filter/    # Query filtering
│   ├── resource/  # REST resource representations
│   ├── association/ # Relationship management
│   └── ...
├── model/         # Domain models
├── task/          # Task management
├── assessment/    # Assessment logic
├── migration/     # Database migrations
└── ...

shared/
├── api/           # Exported REST API types
├── addon/         # Addon framework types
├── binding/       # API client bindings
├── settings/      # Configuration types
└── ...
```

---

## Object-Oriented Design

### Prefer Structs with Methods

**Model behavior using structs with methods** rather than package-level functions. This promotes encapsulation and makes dependencies explicit.

```go
// Good
type ApplicationHandler struct {
    BucketOwner
}

func (h ApplicationHandler) Get(ctx *gin.Context) {
    m := &model.Application{}
    id := h.pk(ctx)
    db := h.preLoad(h.DB(ctx), clause.Associations)
    result := db.First(m, id)
    if result.Error != nil {
        _ = ctx.Error(result.Error)
        return
    }
    // ... rest of implementation
}

// Bad - package functions
func GetApplication(ctx *gin.Context) {
    // implementation
}
```

### Handler Pattern

HTTP handlers should be struct methods that follow this pattern:

```go
type ResourceHandler struct {
    BaseHandler  // Embed common functionality
}

func (h ResourceHandler) Get(ctx *gin.Context) {
    // Implementation
}

func (h ResourceHandler) List(ctx *gin.Context) {
    // Implementation
}

func (h ResourceHandler) Create(ctx *gin.Context) {
    // Implementation
}

func (h ResourceHandler) Update(ctx *gin.Context) {
    // Implementation
}

func (h ResourceHandler) Delete(ctx *gin.Context) {
    // Implementation
}
```

---

## Function Design

### Single Responsibility Principle

**Each function must do one thing and do it well.** Functions should have a single, clear purpose.

```go
// Good - focused functions
func (h *ApplicationHandler) tagMap(ctx *gin.Context, appIds []uint) (mp TagMap, err error) {
    // Only builds and returns the tag map
}

func (h *ApplicationHandler) idMap(ctx *gin.Context, appId uint) (mp IdentityMap, err error) {
    // Only builds and returns the identity map
}

func (h *ApplicationHandler) replaceTags(db *gorm.DB, id uint, r *Application) (appTags []AppTag, err error) {
    // Only replaces tags
}

// Bad - doing multiple things
func (h *ApplicationHandler) getTagsAndIdentitiesAndReplaceTags(ctx *gin.Context, id uint) (tags []Tag, ids []Identity, err error) {
    // Don't combine multiple responsibilities
}
```

### No "And" in Method Names

**Method names containing "and" indicate the method is doing too many things.** Split such methods into multiple focused methods.

```go
// Bad - method name contains "and"
func (h Handler) ValidateAndSave(ctx *gin.Context) {
    // Doing two things
}

func (h Handler) FetchAndTransform(data Data) {
    // Doing two things
}

// Good - separate methods
func (h Handler) Validate(ctx *gin.Context) (err error) {
    // Only validates
    return
}

func (h Handler) Save(ctx *gin.Context) (err error) {
    // Only saves
    return
}
```

### Avoid Long Function Bodies

**Keep function bodies concise.** If a function has multiple distinct code blocks or exceeds ~50 lines, consider decomposing it into smaller functions.

```go
// Bad - long function with multiple blocks
func (h Handler) Process(ctx *gin.Context) {
    // Block 1: Validation (20 lines)
    // ...

    // Block 2: Transformation (30 lines)
    // ...

    // Block 3: Persistence (25 lines)
    // ...

    // Block 4: Notification (15 lines)
    // ...
}

// Good - decomposed
func (h Handler) Process(ctx *gin.Context) (err error) {
    err = h.validate(ctx)
    if err != nil {
        return
    }

    data, err := h.transform(ctx)
    if err != nil {
        return
    }

    err = h.persist(ctx, data)
    if err != nil {
        return
    }

    err = h.notify(ctx)
    return
}

func (h Handler) validate(ctx *gin.Context) (err error) {
    // Focused validation logic
    return
}

func (h Handler) transform(ctx *gin.Context) (data Data, err error) {
    // Focused transformation logic
    return
}

func (h Handler) persist(ctx *gin.Context, data Data) (err error) {
    // Focused persistence logic
    return
}

func (h Handler) notify(ctx *gin.Context) (err error) {
    // Focused notification logic
    return
}
```

---

## Function and Method Standards

### Naming Conventions

### File Names

Use meaningful, descriptive file names that indicate their purpose:

```go
// Good
application.go      // Application handler
task.go            // Task handler
filter.go          // Filtering logic
assessment.go      // Assessment logic
base.go            // Base handler/types

// Bad
helper.go          // Too vague
utils.go           // Too generic
common.go          // Too broad
misc.go            // Indicates poor organization
```

### Variable Names

Use clear, descriptive variable names:

```go
// Good
db := h.DB(ctx)
questResolver, err := assessment.NewQuestionnaireResolver(h.DB(ctx))
appResolver := assessment.NewApplicationResolver(tagResolver, memberResolver, questResolver)

// Acceptable for short scopes
for i := range list {
    m := &list[i]
    // ...
}

// Bad
var x *gorm.DB  // Too vague
var temp string // Non-descriptive
```

### Function Returns

#### Named Return Variables

**Always use named return variables** for function signatures. This provides clarity about what the function returns and enables cleaner error handling patterns.

```go
// Good
func (h *BaseHandler) DB(ctx *gin.Context) (db *gorm.DB) {
    rtx := RichContext(ctx)
    if Log.V(1).Enabled() {
        db = rtx.DB.Debug()
    } else {
        db = rtx.DB
    }
    return
}

// Good
func (h *ApplicationHandler) tagMap(ctx *gin.Context, appIds []uint) (mp TagMap, err error) {
    tagCache := make(map[uint]*model.Tag)
    var tags []*model.Tag
    db := h.DB(ctx)
    err = db.Find(&tags).Error
    if err != nil {
        return
    }
    mp = make(TagMap)
    return
}

// Bad - unnamed returns
func (h *BaseHandler) DB(ctx *gin.Context) *gorm.DB {
    rtx := RichContext(ctx)
    if Log.V(1).Enabled() {
        return rtx.DB.Debug()
    }
    return rtx.DB
}
```

#### Bare Return Statements

When using named return variables, **use bare `return` statements** instead of explicitly returning the values.

```go
// Good
func (h *BaseHandler) pk(ctx *gin.Context) (id uint) {
    s := ctx.Param(ID)
    n, _ := strconv.Atoi(s)
    id = uint(n)
    return
}

// Bad - explicit return with named variables
func (h *BaseHandler) pk(ctx *gin.Context) (id uint) {
    s := ctx.Param(ID)
    n, _ := strconv.Atoi(s)
    id = uint(n)
    return id  // Don't do this
}
```

### Code Documentation and Comments

#### Function and Method Docstrings

**All exported functions and methods must have docstrings** that explain what the function does. Use standard godoc format: start with the function/method name followed by what it does.

**Godoc Format Pattern:**
```
// FunctionName does something and returns something.
```

**Examples:**

```go
// Good - proper godoc format: name + what it does
// DoThing does a thing then returns a thing.
func DoThing() (thing Thing) {
    // Implementation
    return
}

// Good - database method
// DB returns the database client associated with the context.
func (h *BaseHandler) DB(ctx *gin.Context) (db *gorm.DB) {
    rtx := RichContext(ctx)
    if Log.V(1).Enabled() {
        db = rtx.DB.Debug()
    } else {
        db = rtx.DB
    }
    return
}

// Good - multi-line docstring for complex behavior
// WithCount reports the count and sets the X-Total header for pagination.
// Returns an error when count exceeds the limit and is not constrained
// by pagination.
func (h *BaseHandler) WithCount(ctx *gin.Context, count int64) (err error) {
    // Implementation
    return
}

// Good - another example
// tagMap builds and returns a map of application tags indexed by application ID.
func (h *ApplicationHandler) tagMap(ctx *gin.Context, appIds []uint) (mp TagMap, err error) {
    // Implementation
    return
}

// Good - HTTP handler with godoc and swagger annotations
// Get godoc
// @summary Get an application by ID.
// @description Get an application by ID.
// @tags applications
// @produce json
// @success 200 {object} api.Application
// @router /applications/{id} [get]
// @param id path int true "Application ID"
func (h ApplicationHandler) Get(ctx *gin.Context) {
    // Implementation
}

// Bad - doesn't start with function name
// Returns the primary key from context.
func (h *BaseHandler) pk(ctx *gin.Context) (id uint) {
    return
}

// Bad - missing docstring entirely
func (h *BaseHandler) pk(ctx *gin.Context) (id uint) {
    return
}
```

#### Docstring Requirements

- **Public functions/methods**: Must always have docstrings
- **Private functions/methods**: Should have docstrings for non-trivial logic
- **Godoc format**: Start with the function/method name, then describe what it does
  - Example: `// ProcessTask processes the given task and returns the result.`
  - Example: `// Save saves the model to the database.`
  - Example: `// NewClient creates and returns a new API client.`
- **Explain what, not how**: Describe the purpose and behavior, not implementation details
- **Document return values**: Especially error conditions and special cases
- **Include swagger annotations**: For HTTP handler methods (godoc, @summary, @description, etc.)

#### Inline Comments: Document Why, Not What

**Inline comments should be rare and should only explain *why* something is done when it's not obvious from the code itself.** The code should be self-documenting through clear naming and structure.

```go
// Good - explains WHY, provides important context
func (h ApplicationHandler) AssessmentCreate(ctx *gin.Context) {
    m := r.Model()
    m.CreateUser = h.CurrentUser(ctx)

    // If sections aren't empty that indicates this assessment is being
    // created "as-is" and should not have its sections populated or autofilled.
    newAssessment := false
    if len(m.Sections) == 0 {
        m.Sections = q.Sections
        assessment.PrepareForApplication(resolver, application, m)
        newAssessment = true
    }
    // ... rest of implementation
}

// Bad - stating the obvious
func (h ApplicationHandler) Delete(ctx *gin.Context) {
    // Get the application from the database
    result := h.DB(ctx).First(m, id)
    // Delete the application
    result = h.DB(ctx).Delete(m)
    // Return no content status
    h.Status(ctx, http.StatusNoContent)
}

// Good - self-documenting, no comments needed
func (h ApplicationHandler) Delete(ctx *gin.Context) {
    result := h.DB(ctx).First(m, id)
    if result.Error != nil {
        _ = ctx.Error(result.Error)
        return
    }
    result = h.DB(ctx).Delete(m)
    if result.Error != nil {
        _ = ctx.Error(result.Error)
        return
    }
    h.Status(ctx, http.StatusNoContent)
}
```

#### When to Write Inline Comments

**DO write inline comments when:**

- Explaining a non-obvious business rule or requirement
- Documenting a workaround for a bug in a dependency
- Clarifying complex algorithms or calculations
- Explaining why a particular approach was chosen over alternatives
- Warning about edge cases or gotchas
- Documenting assumptions or constraints

```go
// Good - explains WHY a workaround is needed
func (h Handler) ProcessData(data []byte) (err error) {
    // PostgreSQL has a limit of 65535 parameters in a query.
    // Batch the inserts to stay under this limit.
    batchSize := 1000
    for i := 0; i < len(data); i += batchSize {
        // Process batch
    }
    return
}

// Good - documents a business rule
func (h Handler) CalculateRisk(app *Application) (risk string) {
    // Risk is inherited from archetypes only when the application
    // has no direct assessment. This preserves explicit user choices.
    if app.HasAssessment() {
        risk = app.Assessment.Risk
    } else {
        risk = app.InheritedRisk()
    }
    return
}

// Good - explains a non-intuitive algorithm choice
func (h Handler) SortTasks(tasks []Task) {
    // Use stable sort to preserve creation order for tasks with equal priority.
    // This ensures deterministic ordering in the UI.
    sort.SliceStable(tasks, func(i, j int) bool {
        return tasks[i].Priority > tasks[j].Priority
    })
}
```

**DON'T write inline comments when:**

- The comment just repeats what the code obviously does
- Proper naming would make the comment unnecessary
- The comment describes what a standard library function does
- The comment could be eliminated by extracting a well-named function

#### Comment Style Guidelines

- Use `//` for all comments (not `/* */` except for package documentation)
- Start comments with a capital letter
- End with a period for complete sentences
- Keep comments concise and focused
- Update comments when code changes (stale comments are worse than no comments)

```go
// Good comment style
// This function processes items in parallel to improve performance.
// It spawns up to maxWorkers goroutines based on available CPU cores.
func (h Handler) ProcessParallel(items []Item) (err error) {
    // Implementation
    return
}

// Bad comment style
// this processes items
func (h Handler) ProcessParallel(items []Item) (err error) {
    // Implementation
    return
}
```

#### TODO Comments

Use TODO comments for temporary code or future improvements, but include context:

```go
// Good - TODO with context
// TODO(username): Optimize this query using a materialized view once we upgrade to PostgreSQL 14.
func (h Handler) ComplexQuery() (results []Result, err error) {
    // Current implementation
    return
}

// Bad - vague TODO
// TODO: make this better
func (h Handler) ComplexQuery() (results []Result, err error) {
    return
}
```

#### Summary: Comment Philosophy

- **Code tells you HOW, comments tell you WHY**
- **Write self-documenting code through clear naming and structure**
- **Add comments only when they provide value beyond what the code itself expresses**
- **Every public API must be documented**
- **Inline comments should be the exception, not the rule**

---

## Testing

### Test File Organization

**Prefer a single test file per package** named `pkg_test.go` unless the number of test functions makes this impractical.

```bash
# Good - single test file for package
internal/api/filter/
├── filter.go
├── pkg.go
└── pkg_test.go      # All filter package tests

# Also acceptable - using package name
internal/api/filter/
├── filter.go
└── filter_test.go

# Good - when tests become too large, split by domain/feature
internal/api/
├── application.go
├── pkg_test.go           # Core tests
├── facts_test.go         # Separate file for facts-specific tests
└── task.go

# Bad - unnecessary proliferation of test files
internal/model/
├── application.go
├── create_test.go   # Don't split by operation
├── update_test.go   # unless necessary
├── delete_test.go
└── get_test.go
```

**When to use multiple test files:**
- Package has a large number of test functions (>20-30)
- Tests naturally group into distinct feature areas
- Integration tests vs unit tests separation

**Naming convention:**
- Primary test file: `pkg_test.go` (preferred, avoids redundant package name)
- Alternative: `<package>_test.go` (acceptable)
- Additional test files: `<feature>_test.go` (e.g., `facts_test.go`, `validation_test.go`)

### Use Gomega for Assertions

**All tests must use Gomega for assertions** instead of basic Go testing comparisons. Gomega provides better error messages and more expressive assertions.

```go
// Good - using Gomega
func TestLexer(t *testing.T) {
    g := gomega.NewGomegaWithT(t)
    lexer := Lexer{}
    err := lexer.With("name:elmer,age:20")
    g.Expect(err).To(gomega.BeNil())
    g.Expect(lexer.tokens).To(gomega.Equal(
        []Token{
            {Kind: LITERAL, Value: "name"},
            {Kind: OPERATOR, Value: string(COLON)},
            {Kind: LITERAL, Value: "elmer"},
        }))
}

// Bad - basic testing
func TestLexer(t *testing.T) {
    lexer := Lexer{}
    err := lexer.With("name:elmer,age:20")
    if err != nil {
        t.Errorf("Expected no error, got %v", err)
    }
    if len(lexer.tokens) != 3 {
        t.Errorf("Expected 3 tokens, got %d", len(lexer.tokens))
    }
}
```

### Common Gomega Assertions

```go
g := gomega.NewGomegaWithT(t)

// Nil checks
g.Expect(err).To(gomega.BeNil())
g.Expect(err).NotTo(gomega.BeNil())

// Equality
g.Expect(actual).To(gomega.Equal(expected))

// Boolean checks
g.Expect(value).To(gomega.BeTrue())
g.Expect(value).To(gomega.BeFalse())

// Collection checks
g.Expect(list).To(gomega.HaveLen(5))
g.Expect(list).To(gomega.ContainElement(item))
g.Expect(list).To(gomega.BeEmpty())

// String matching
g.Expect(str).To(gomega.ContainSubstring("substring"))
g.Expect(str).To(gomega.MatchRegexp("pattern"))
```

### Use cmp.Eq() for Complex Struct Comparisons

**For comparing complex structs and objects, use `test/cmp.Eq()`** instead of Gomega's `Equal()`. Benefits:
- Deep equality checking with detailed diff reports
- Ignore specific fields (timestamps, auto-generated IDs)
- Clear, readable diff output when comparisons fail

```go
import (
    "testing"
    "github.com/konveyor/tackle2-hub/test/cmp"
    . "github.com/onsi/gomega"
)

// Basic usage - ignore specific fields
func TestTrackerUpdate(t *testing.T) {
    g := NewGomegaWithT(t)
    tracker := &Tracker{ID: 1, Name: "Jira", URL: "https://jira.example.com"}

    err := client.Create(tracker)
    g.Expect(err).To(BeNil())

    retrieved, err := client.Get(tracker.ID)
    g.Expect(err).To(BeNil())

    // Ignore LastUpdated timestamp
    eq, report := cmp.Eq(tracker, retrieved, "LastUpdated")
    g.Expect(eq).To(BeTrue(), report)
}

// Advanced - multiple ignored fields and fluent API
func TestApplicationComparison(t *testing.T) {
    g := NewGomegaWithT(t)
    expected := &Application{Name: "MyApp", Tags: []Tag{{ID: 1, Name: "Java"}}}
    actual := &Application{Name: "MyApp", Tags: []Tag{{ID: 1, Name: "Java"}},
        CreateUser: "admin", LastUpdated: time.Now()}

    // Ignore audit fields
    eq, report := cmp.Eq(expected, actual, "CreateUser", "UpdateUser", "LastUpdated")
    g.Expect(eq).To(BeTrue(), report)

    // Or fluent API with sorting
    eq, report = cmp.New().Sort(sort.ById, []Task{}).Ignore("Created").Eq(expected, actual)
    g.Expect(eq).To(BeTrue(), report)
}
```

**When to use `cmp.Eq()`:**
- REST API resources with audit fields or timestamps
- Database models with auto-generated fields
- Nested structures where specific fields should be ignored
- Need detailed diff output for debugging

**When `gomega.Equal()` is sufficient:**
- Primitive types (strings, ints, bools)
- Simple slices or maps
- Exact equality required for all fields

### Test Structure

```go
func TestFeatureName(t *testing.T) {
    g := gomega.NewGomegaWithT(t)

    // Setup
    resource := &Resource{Name: "test"}

    // Execute
    err := resource.Process()

    // Assert
    g.Expect(err).To(gomega.BeNil())
    g.Expect(resource.Status).To(gomega.Equal("processed"))
}
```

---

## Error Handling

### Error Returns

Always include error returns in function signatures when operations can fail:

```go
// Good
func (h *Handler) Process(ctx *gin.Context) (err error) {
    data, err := h.fetch(ctx)
    if err != nil {
        return
    }

    err = h.validate(data)
    if err != nil {
        return
    }

    err = h.save(data)
    return
}

// Bad - ignoring potential errors
func (h *Handler) Process(ctx *gin.Context) {
    data := h.fetch(ctx)  // What if this fails?
    h.validate(data)
    h.save(data)
}
```

### Error Handling in HTTP Handlers

In Gin HTTP handlers, use the context error mechanism:

```go
func (h ApplicationHandler) Get(ctx *gin.Context) {
    m := &model.Application{}
    id := h.pk(ctx)
    db := h.preLoad(h.DB(ctx), clause.Associations)
    result := db.First(m, id)
    if result.Error != nil {
        _ = ctx.Error(result.Error)
        return
    }

    tagMap, err := h.tagMap(ctx, []uint{id})
    if err != nil {
        _ = ctx.Error(err)
        return
    }

    h.Respond(ctx, http.StatusOK, r)
}
```

### Error Reporting

Errors should either be handled or wrapped and returned. Errors _caught_ at the
root of the stack must be logged. Logging errors everywhere they are _caught_ assumes
they cannot be handled by the caller. This is usually a bad assumption and results
in errors begin reported then handled and/or reported multiple times. In both cases
the log ends up with noise and Red Herrings, which results in concerned users and
bugs reported for routine, anticipated (non)errors.

The liberr.Wrap() captures the stack trace and context and returns
wrapped error.  Both the trace and context is reported by the logger.
Having this information about WHERE the error originated is much more
useful than knowing where the error was _caught_.

To prevent multiple (unnecessary) wrapping, errors should only be wrapped at the _edge_. This
means only when returned by external packages.  Assume an error returned by a
method within the project has already been wrapped.


```go
import (
    liberr "github.com/jortel/go-utils/error"
)
```

```go

func (m *Manager) DoThing() (err error) {
    f, err = os.Open("")
    if err != nil {
        err = liberr.Wrap(err)
        return
    }
}
```

---

## Git Commits

### Signed Commits Required

**All commits MUST be signed off** using the `-s` flag. This adds a "Signed-off-by" line certifying you have the right to submit the code under the project's license.

```bash
# Required - always use -s flag
git commit -s -m "Fix application tag filtering"

# Recommended - combine with -a and -m
git commit -sam "Fix application tag filtering"

# Bad - missing -s flag
git commit -am "Fix application tag filtering"
```

The `-s` flag adds a line like this to your commit:
```
Signed-off-by: Your Name <your.email@example.com>
```

---

## Summary

Following these standards ensures:
- **Consistency** across the codebase
- **Readability** for all contributors
- **Maintainability** over time
- **Quality** through good design practices

When in doubt, look at existing code in `internal/api/`, `internal/model/`, and `internal/task/` for examples of these patterns in practice.

## References

- See `internal/api/application.go` for handler examples
- See `internal/api/base.go` for base handler patterns
- See `internal/api/filter/filter_test.go` for Gomega test examples
- See `test/cmp/eq_test.go` for cmp.Eq() usage examples
- See `test/binding/tracker_test.go` for cmp.Eq() in integration tests
- See `shared/api/` for exported REST API types
- See `shared/binding/` for client binding implementations
- See `internal/assessment/` for domain logic organization
