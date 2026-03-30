# API Endpoint Design Patterns and Conventions

This document describes the design patterns, conventions, and standards for implementing REST API endpoints in the Tackle2 Hub.

## Table of Contents

- [Complete File Structure](#complete-file-structure)
- [Adding a New Endpoint - Complete Example](#adding-a-new-endpoint---complete-example)
- [Handler Architecture](#handler-architecture)
- [Handler Implementation Patterns](#handler-implementation-patterns)
- [Route Registration](#route-registration)
- [Request/Response Flow](#requestresponse-flow)
- [Resource Conversion](#resource-conversion)
- [Error Handling](#error-handling)
- [Authentication and Authorization](#authentication-and-authorization)
- [Associations and Relationships](#associations-and-relationships)
- [Sub-Resources](#sub-resources)
- [Pagination and Filtering](#pagination-and-filtering)
- [Testing API Handlers](#testing-api-handlers)

---

## Complete File Structure

When adding a new REST API endpoint, you will need to create or modify files in multiple locations:

```text
tackle2-hub/
├── internal/
│   ├── api/
│   │   ├── resource.go              # Handler implementation (CREATE THIS)
│   │   ├── pkg.go                   # Register handler in All() (MODIFY THIS)
│   │   └── resource/
│   │       └── resource.go          # Resource converter (CREATE THIS)
│   ├── model/
│   │   └── pkg.go                   # Re-export model type (MODIFY THIS)
│   └── migration/
│       └── v{N}/                    # Current migration version (e.g., v21)
│           └── model/
│               └── application.go   # Model definition (MODIFY THIS)
│                   or core.go       # or other appropriate file
└── shared/
    └── api/
        ├── application.go           # REST API type definition (MODIFY THIS)
        │   or core.go               # or other appropriate file
        └── pkg.go                   # Route constants (MODIFY THIS)
```

### File Locations Explained

1. **Model Definition** (`internal/migration/v{N}/model/*.go`)
   - Database model with GORM tags
   - Latest version is current (e.g., v21)
   - Group related models in same file (e.g., `application.go` for app-related models)

2. **REST API Type** (`shared/api/*.go`)
   - Client-facing type definition
   - JSON/YAML struct tags
   - Group related types in same file (e.g., `application.go`)
   - NEVER import from `internal/` package

3. **Route Constants** (`shared/api/pkg.go`)
   - Define route path constants
   - Organized by resource type with comments

4. **Handler** (`internal/api/*.go`)
   - HTTP handler methods (Get, List, Create, Update, Delete)
   - One file per resource
   - Filename matches resource (lowercase, e.g., `jobfunction.go`)

5. **Resource Converter** (`internal/api/resource/*.go`)
   - Converts between model and REST API type
   - One file per resource
   - Filename matches resource

6. **Handler Registration** (`internal/api/pkg.go`)
   - Add handler to `All()` function

---

## Adding a New Endpoint - Complete Example

This example shows **every file** you need to create or modify to add a new "Widget" endpoint.

### Step 1: Define the Model

**File:** `internal/migration/v21/model/core.go` (or appropriate file)

```go
// Widget model for database.
type Widget struct {
	Model
	Name        string `gorm:"index;unique;not null"`
	Description string
	OwnerID     *uint
	Owner       *Stakeholder
}
```

**Notes:**
- Embed `Model` for ID, timestamps, audit fields
- Use GORM tags for constraints
- Foreign keys as pointer to uint
- Associations as pointer to model type

**File:** `internal/model/pkg.go` (re-export the model)

```go
type Widget = model.Widget
```

**Notes:**
- Add type alias in alphabetical order
- This makes `model.Widget` available to handlers

### Step 2: Define REST API Type

**File:** `shared/api/core.go` (or create new file)

```go
// Widget REST resource.
type Widget struct {
	Resource    `yaml:",inline"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	Owner       *Ref   `json:"owner"`
}
```

**Notes:**
- Embed `Resource` for ID, timestamps, audit fields
- Use JSON tags for field names
- Use `binding` tags for validation
- Foreign keys as `*Ref` (reference to related resource)

### Step 3: Define Route Constants

**File:** `shared/api/pkg.go`

```go
// Routes - Widgets
const (
	WidgetsRoute = "/widgets"
	WidgetRoute  = WidgetsRoute + "/:" + ID
)
```

**Notes:**
- Add in appropriate alphabetical location
- Use comment header to group routes
- Plural for collection, singular for item

### Step 4: Create Resource Converter

**File:** `internal/api/resource/widget.go`

```go
package resource

import (
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
)

// Widget REST resource.
type Widget api.Widget

// With converts model to REST resource.
func (r *Widget) With(m *model.Widget) {
	baseWith(&r.Resource, &m.Model)
	r.Name = m.Name
	r.Description = m.Description
	r.Owner = refPtr(m.OwnerID, m.Owner)
}

// Model converts REST resource to model.
func (r *Widget) Model() (m *model.Widget) {
	m = &model.Widget{
		Name:        r.Name,
		Description: r.Description,
	}
	m.ID = r.ID
	if r.Owner != nil {
		m.OwnerID = &r.Owner.ID
	}
	return
}
```

**Notes:**
- Type alias to `api.Widget`
- `With()` for model → resource (outbound)
- `Model()` for resource → model (inbound)
- Use `baseWith()` for common fields
- Use `refPtr()` for optional foreign key references
- Use `ref()` for required foreign key references

### Step 5: Create Handler

**File:** `internal/api/widget.go`

```go
package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/konveyor/tackle2-hub/internal/api/resource"
	"github.com/konveyor/tackle2-hub/internal/model"
	"github.com/konveyor/tackle2-hub/shared/api"
	"gorm.io/gorm/clause"
)

// WidgetHandler handles widget routes.
type WidgetHandler struct {
	BaseHandler
}

// AddRoutes adds routes.
func (h WidgetHandler) AddRoutes(e *gin.Engine) {
	routeGroup := e.Group("/")
	routeGroup.Use(Required("widgets"))
	routeGroup.GET(api.WidgetsRoute, h.List)
	routeGroup.GET(api.WidgetsRoute+"/", h.List)
	routeGroup.POST(api.WidgetsRoute, h.Create)
	routeGroup.GET(api.WidgetRoute, h.Get)
	routeGroup.PUT(api.WidgetRoute, h.Update)
	routeGroup.DELETE(api.WidgetRoute, h.Delete)
}

// Get godoc
// @summary Get a widget by ID.
// @description Get a widget by ID.
// @tags widgets
// @produce json
// @success 200 {object} api.Widget
// @router /widgets/{id} [get]
// @param id path int true "Widget ID"
func (h WidgetHandler) Get(ctx *gin.Context) {
	m := &model.Widget{}
	id := h.pk(ctx)
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	r := Widget{}
	r.With(m)
	h.Respond(ctx, http.StatusOK, r)
}

// List godoc
// @summary List all widgets.
// @description List all widgets.
// @tags widgets
// @produce json
// @success 200 {object} []api.Widget
// @router /widgets [get]
func (h WidgetHandler) List(ctx *gin.Context) {
	var list []model.Widget
	db := h.preLoad(h.DB(ctx), clause.Associations)
	err := db.Find(&list).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	resources := []Widget{}
	for i := range list {
		r := Widget{}
		r.With(&list[i])
		resources = append(resources, r)
	}

	h.Respond(ctx, http.StatusOK, resources)
}

// Create godoc
// @summary Create a widget.
// @description Create a widget.
// @tags widgets
// @accept json
// @produce json
// @success 201 {object} api.Widget
// @router /widgets [post]
// @param widget body api.Widget true "Widget data"
func (h WidgetHandler) Create(ctx *gin.Context) {
	r := &Widget{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.CreateUser = h.CurrentUser(ctx)
	err = h.DB(ctx).Create(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	r.With(m)

	h.Respond(ctx, http.StatusCreated, r)
}

// Update godoc
// @summary Update a widget.
// @description Update a widget.
// @tags widgets
// @accept json
// @success 204
// @router /widgets/{id} [put]
// @param id path int true "Widget ID"
// @param widget body api.Widget true "Widget data"
func (h WidgetHandler) Update(ctx *gin.Context) {
	id := h.pk(ctx)
	r := &Widget{}
	err := h.Bind(ctx, r)
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	m := r.Model()
	m.ID = id
	m.UpdateUser = h.CurrentUser(ctx)
	db := h.DB(ctx).Model(m)
	db = db.Omit(clause.Associations)
	err = db.Save(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Delete godoc
// @summary Delete a widget.
// @description Delete a widget.
// @tags widgets
// @success 204
// @router /widgets/{id} [delete]
// @param id path int true "Widget ID"
func (h WidgetHandler) Delete(ctx *gin.Context) {
	id := h.pk(ctx)
	m := &model.Widget{}
	err := h.DB(ctx).First(m, id).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}
	err = h.DB(ctx).Delete(m).Error
	if err != nil {
		_ = ctx.Error(err)
		return
	}

	h.Status(ctx, http.StatusNoContent)
}

// Widget REST resource.
type Widget = resource.Widget
```

**Notes:**
- Filename is lowercase resource name
- Struct embeds `BaseHandler`
- `AddRoutes()` registers all routes
- Both `/widgets` and `/widgets/` for List
- Type alias at bottom: `type Widget = resource.Widget`

### Step 6: Register Handler

**File:** `internal/api/pkg.go`

```go
func All() []Handler {
	return []Handler{
		&AddonHandler{},
		// ... existing handlers ...
		&WidgetHandler{},  // Add this line in alphabetical order
		// ... more handlers ...
	}
}
```

### Verification Checklist

After implementing, verify:

**Files Created/Modified:**
- [ ] Model defined in `internal/migration/v{N}/model/*.go`
- [ ] Model re-exported in `internal/model/pkg.go`
- [ ] REST API type defined in `shared/api/*.go`
- [ ] Route constants defined in `shared/api/pkg.go`
- [ ] Resource converter created in `internal/api/resource/widget.go`
- [ ] Handler created in `internal/api/widget.go`
- [ ] Handler registered in `internal/api/pkg.go`

**Implementation:**
- [ ] All CRUD methods implemented (Get, List, Create, Update, Delete)
- [ ] Godoc comments on all methods
- [ ] Swagger annotations on all methods
- [ ] Proper middleware (`Required`, `Transaction`)
- [ ] Correct status codes (200, 201, 204)
- [ ] Using `ctx.Error()` for all errors
- [ ] Setting `CreateUser`/`UpdateUser`
- [ ] Using `Omit(clause.Associations)` in Update

---

## Handler Architecture

### Handler Structure

**Every REST resource is managed by a handler struct** that implements the `Handler` interface:

```go
// Handler interface - all handlers must implement this
type Handler interface {
    AddRoutes(e *gin.Engine)
}

// Standard handler pattern
type ResourceHandler struct {
    BaseHandler  // Embed BaseHandler for common functionality
}

// Handlers with bucket support
type ApplicationHandler struct {
    BucketOwner  // Embed BucketOwner (which embeds BaseHandler)
}
```

**Key Principles:**
- One handler per REST resource
- Embed `BaseHandler` for standard resources
- Embed `BucketOwner` for resources that manage file storage
- All handlers registered in `pkg.go:All()` function

### Handler Registration

All handlers are instantiated and registered in `pkg.go`:

```go
func All() []Handler {
    return []Handler{
        &ApplicationHandler{},
        &BusinessServiceHandler{},
        &StakeholderHandler{},
        &TagHandler{},
        // ... other handlers
    }
}
```

---

## Handler Implementation Patterns

### Standard CRUD Operations

Every handler implements standard CRUD methods following this exact pattern:

#### Get - Retrieve Single Resource

```go
// Get godoc
// @summary Get a resource by ID.
// @description Get a resource by ID.
// @tags resources
// @produce json
// @success 200 {object} api.Resource
// @router /resources/{id} [get]
// @param id path int true "Resource ID"
func (h ResourceHandler) Get(ctx *gin.Context) {
    m := &model.Resource{}
    id := h.pk(ctx)
    db := h.preLoad(h.DB(ctx), clause.Associations)
    err := db.First(m, id).Error
    if err != nil {
        _ = ctx.Error(err)
        return
    }

    r := Resource{}
    r.With(m)
    h.Respond(ctx, http.StatusOK, r)
}
```

**Pattern:**
1. Get ID from URL using `h.pk(ctx)`
2. Get DB client with `h.DB(ctx)`
3. Use `h.preLoad()` for associations
4. Fetch model with `db.First()`
5. Convert error with `ctx.Error()` if failed
6. Convert model to REST resource with `r.With(m)`
7. Respond with 200 OK

#### List - Retrieve All Resources

```go
// List godoc
// @summary List all resources.
// @description List all resources.
// @tags resources
// @produce json
// @success 200 {object} []api.Resource
// @router /resources [get]
func (h ResourceHandler) List(ctx *gin.Context) {
    var list []model.Resource
    db := h.preLoad(h.DB(ctx), clause.Associations)
    err := db.Find(&list).Error
    if err != nil {
        _ = ctx.Error(err)
        return
    }
    resources := []Resource{}
    for i := range list {
        r := Resource{}
        r.With(&list[i])
        resources = append(resources, r)
    }

    h.Respond(ctx, http.StatusOK, resources)
}
```

**Pattern:**
1. Declare slice for models
2. Fetch all with `db.Find()`
3. Convert each model to REST resource
4. Respond with 200 OK and slice

#### Create - Create New Resource

```go
// Create godoc
// @summary Create a resource.
// @description Create a resource.
// @tags resources
// @accept json
// @produce json
// @success 201 {object} api.Resource
// @router /resources [post]
// @param resource body api.Resource true "Resource data"
func (h ResourceHandler) Create(ctx *gin.Context) {
    r := &Resource{}
    err := h.Bind(ctx, r)
    if err != nil {
        _ = ctx.Error(err)
        return
    }
    m := r.Model()
    m.CreateUser = h.CurrentUser(ctx)
    err = h.DB(ctx).Create(m).Error
    if err != nil {
        _ = ctx.Error(err)
        return
    }
    r.With(m)

    h.Respond(ctx, http.StatusCreated, r)
}
```

**Pattern:**
1. Bind request body with `h.Bind(ctx, r)`
2. Convert resource to model with `r.Model()`
3. Set `CreateUser` from auth token
4. Create in database with `db.Create()`
5. Convert model back to resource (gets ID, timestamps)
6. Respond with 201 Created

#### Update - Update Existing Resource

```go
// Update godoc
// @summary Update a resource.
// @description Update a resource.
// @tags resources
// @accept json
// @success 204
// @router /resources/{id} [put]
// @param id path int true "Resource ID"
// @param resource body api.Resource true "Resource data"
func (h ResourceHandler) Update(ctx *gin.Context) {
    id := h.pk(ctx)
    r := &Resource{}
    err := h.Bind(ctx, r)
    if err != nil {
        _ = ctx.Error(err)
        return
    }
    m := r.Model()
    m.ID = id
    m.UpdateUser = h.CurrentUser(ctx)
    db := h.DB(ctx).Model(m)
    db = db.Omit(clause.Associations)
    err = db.Save(m).Error
    if err != nil {
        _ = ctx.Error(err)
        return
    }

    h.Status(ctx, http.StatusNoContent)
}
```

**Pattern:**
1. Get ID from URL
2. Bind request body
3. Convert to model and set ID
4. Set `UpdateUser` from auth token
5. Use `Omit(clause.Associations)` to prevent auto-save of associations
6. Save with `db.Save()`
7. Respond with 204 No Content

#### Delete - Delete Resource

```go
// Delete godoc
// @summary Delete a resource.
// @description Delete a resource.
// @tags resources
// @success 204
// @router /resources/{id} [delete]
// @param id path int true "Resource ID"
func (h ResourceHandler) Delete(ctx *gin.Context) {
    id := h.pk(ctx)
    m := &model.Resource{}
    err := h.DB(ctx).First(m, id).Error
    if err != nil {
        _ = ctx.Error(err)
        return
    }
    err = h.DB(ctx).Delete(m).Error
    if err != nil {
        _ = ctx.Error(err)
        return
    }

    h.Status(ctx, http.StatusNoContent)
}
```

**Pattern:**
1. Get ID from URL
2. Fetch existing record (to verify it exists)
3. Delete with `db.Delete()`
4. Respond with 204 No Content

### BaseHandler Helper Methods

The `BaseHandler` provides essential helper methods:

```go
// Database access
db := h.DB(ctx)                    // Get DB client from context
db = h.preLoad(db, "Association")  // Configure GORM preloading

// Request handling
id := h.pk(ctx)                    // Extract ID from URL parameter
err := h.Bind(ctx, &resource)      // Bind JSON/YAML request body
err = h.Validate(ctx, &resource)   // Validate struct field tags

// Response handling
h.Respond(ctx, http.StatusOK, resource)      // Set status and body
h.Status(ctx, http.StatusNoContent)          // Set status only

// Authentication/Authorization
user := h.CurrentUser(ctx)                   // Get username from token
hasScope := h.HasScope(ctx, "resource:read") // Check scope

// Security
err = h.Encrypt(model)             // Encrypt sensitive fields
err = h.Decrypt(ctx, model)        // Decrypt (if authorized)

// Other utilities
accepted := h.Accepted(ctx, "application/json") // Check Accept header
h.Attachment(ctx, "filename.tar")               // Set download header
client := h.Client(ctx)                         // Get k8s client
```

### Response Status Codes

**Use these standard HTTP status codes:**

```go
// Success responses
http.StatusOK          // 200 - GET (single resource or list)
http.StatusCreated     // 201 - POST (return created resource)
http.StatusNoContent   // 204 - PUT, DELETE (no body returned)

// Error responses (handled by ErrorHandler middleware)
http.StatusBadRequest          // 400 - Validation errors, bad input
http.StatusUnauthorized        // 401 - Authentication failed
http.StatusForbidden           // 403 - Authorization failed
http.StatusNotFound            // 404 - Resource not found
http.StatusConflict            // 409 - Constraint violation, cyclic dependency
http.StatusInternalServerError // 500 - Unexpected errors
http.StatusServiceUnavailable  // 503 - External service unavailable
```

**Never return errors inline** - always use `ctx.Error(err)` and let the `ErrorHandler` middleware handle them.

---

## Route Registration

### Standard Route Pattern

```go
func (h ResourceHandler) AddRoutes(e *gin.Engine) {
    routeGroup := e.Group("/")
    routeGroup.Use(Required("resources"))
    routeGroup.GET(api.ResourcesRoute, h.List)
    routeGroup.GET(api.ResourcesRoute+"/", h.List)
    routeGroup.POST(api.ResourcesRoute, h.Create)
    routeGroup.GET(api.ResourceRoute, h.Get)
    routeGroup.PUT(api.ResourceRoute, h.Update)
    routeGroup.DELETE(api.ResourceRoute, h.Delete)
}
```

**Key points:**
- Create route group with `e.Group("/")`
- Apply middleware with `routeGroup.Use()`
- `Required(scope)` enforces authentication/authorization (always required)
- `Transaction` is optional - add when managing associations (see Middleware Usage)
- Register both `/resources` and `/resources/` for List
- Routes defined in `shared/api/route.go`

### Middleware Usage

**Required Middleware:**
- `Required(scope)` - Authentication and authorization (always required)

**Transaction Middleware:**
- `Transaction` - Wraps POST/PUT/PATCH/DELETE in DB transactions
- **When to use:** Only when handlers need atomic multi-step operations
- **Examples:** Managing associations explicitly, multiple related updates
- **Default:** Most simple CRUD handlers don't need it

```go
// Simple handler - no Transaction needed
func (h TagHandler) AddRoutes(e *gin.Engine) {
    routeGroup := e.Group("/")
    routeGroup.Use(Required("tags"))
    routeGroup.GET(api.TagsRoute, h.List)
    routeGroup.POST(api.TagsRoute, h.Create)
    // ... other routes
}

// Complex handler - Transaction for association management
func (h StakeholderHandler) AddRoutes(e *gin.Engine) {
    routeGroup := e.Group("/")
    routeGroup.Use(Required("stakeholders"), Transaction)
    routeGroup.GET(api.StakeholdersRoute, h.List)
    routeGroup.POST(api.StakeholdersRoute, h.Create)
    // ... other routes
}

// Multiple route groups for different scopes
routeGroup := e.Group("/")
routeGroup.Use(Required("applications"), Transaction)
routeGroup.GET(api.ApplicationsRoute, h.List)
routeGroup.POST(api.ApplicationsRoute, h.Create)

routeGroup = e.Group("/")
routeGroup.Use(Required("applications.facts"), Transaction)
routeGroup.GET(api.ApplicationFactsRoute, h.FactGet)
routeGroup.POST(api.ApplicationFactsRoute, h.FactCreate)
```

---

## Request/Response Flow

### Request Processing Flow

1. **Gin receives HTTP request**
2. **Middleware chain executes:**
   - `Required(scope)` - Authenticate and authorize
   - `Transaction` - Start DB transaction (POST/PUT/PATCH/DELETE only)
3. **Handler method executes:**
   - Extract parameters
   - Bind request body (if applicable)
   - Access database via `h.DB(ctx)`
   - Perform business logic
   - Set response via `h.Respond()` or `h.Status()`
4. **Middleware chain completes:**
   - `Transaction` - Commit or rollback based on errors
   - `ErrorHandler` - Convert errors to HTTP responses
   - `Render` - Serialize response body (JSON/YAML)
5. **Response sent to client**

### Context Flow

The `Context` (RichContext) carries state through the request:

```go
type Context struct {
    *gin.Context
    DB           *gorm.DB       // Database client
    User         string         // Authenticated user
    Scope        struct {       // Authorization scopes
        Granted  []auth.Scope
        Required []string
    }
    Client       client.Client  // Kubernetes client
    Response     Response       // Response to send
    TaskManager  *tasking.Manager
}
```

Access with: `rtx := RichContext(ctx)`

---

## Resource Conversion

### Resource Pattern

Resources in `internal/api/resource/` convert between models and REST API types:

```go
package resource

import (
    "github.com/konveyor/tackle2-hub/internal/model"
    "github.com/konveyor/tackle2-hub/shared/api"
)

// Resource type - alias to shared/api type
type BusinessService api.BusinessService

// With converts model to REST resource
func (r *BusinessService) With(m *model.BusinessService) {
    baseWith(&r.Resource, &m.Model)  // Copy ID, timestamps, audit fields
    r.Name = m.Name
    r.Description = m.Description
    r.Stakeholder = refPtr(m.StakeholderID, m.Stakeholder)
}

// Model converts REST resource to model
func (r *BusinessService) Model() (m *model.BusinessService) {
    m = &model.BusinessService{
        Name:        r.Name,
        Description: r.Description,
    }
    m.ID = r.ID
    if r.Stakeholder != nil {
        m.StakeholderID = &r.Stakeholder.ID
    }
    return
}
```

### Helper Functions

```go
// Convert model audit fields to resource
baseWith(&r.Resource, &m.Model)

// Create reference pointer from ID and model
ref := refPtr(m.ForeignKeyID, m.Association)

// Create reference from ID and model (not pointer)
ref := ref(id, model)

// Extract ID pointer from reference
id := idPtr(ref)
```

### Reference Pattern

Foreign key relationships are represented as `Ref` in REST API:

```go
type Ref struct {
    ID   uint   `json:"id"`
    Name string `json:"name"`
}

// In resource
type BusinessService struct {
    Resource
    Name         string `json:"name"`
    Stakeholder  *Ref   `json:"stakeholder"`  // Pointer = optional
}

// In model
type BusinessService struct {
    Model
    Name           string
    StakeholderID  *uint         // Foreign key
    Stakeholder    *Stakeholder  // Association (loaded with preload)
}

// Conversion
r.Stakeholder = refPtr(m.StakeholderID, m.Stakeholder)
```

---

## Error Handling

### Error Propagation Pattern

**CRITICAL: Never handle errors inline in handlers** - always delegate to ErrorHandler middleware:

```go
// Good - delegate to ErrorHandler
err := h.DB(ctx).First(m, id).Error
if err != nil {
    _ = ctx.Error(err)  // Use _ since return value is always nil
    return
}

// Bad - handling errors inline
err := h.DB(ctx).First(m, id).Error
if err != nil {
    ctx.JSON(http.StatusNotFound, gin.H{"error": "not found"})
    return
}
```

### Custom Error Types

Define custom errors in `error.go`:

```go
type BadRequestError struct {
    Reason string
}

func (r *BadRequestError) Error() string {
    return r.Reason
}

func (r *BadRequestError) Is(err error) (matched bool) {
    var target *BadRequestError
    matched = errors.As(err, &target)
    return
}
```

Usage:

```go
if invalidInput {
    err := &BadRequestError{Reason: "Invalid field format"}
    _ = ctx.Error(err)
    return
}
```

### ErrorHandler Middleware

The `ErrorHandler` middleware maps error types to HTTP status codes:

```go
func ErrorHandler() gin.HandlerFunc {
    return func(ctx *gin.Context) {
        ctx.Next()

        if len(ctx.Errors) == 0 {
            return
        }

        err := ctx.Errors[0]

        // Map error types to status codes
        if errors.Is(err, &BadRequestError{}) {
            rtx.Respond(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        if errors.Is(err, gorm.ErrRecordNotFound) {
            // DELETE returns 204 even if not found
            if ctx.Request.Method == http.MethodDelete {
                rtx.Status(http.StatusNoContent)
                return
            }
            rtx.Respond(http.StatusNotFound, gin.H{"error": err.Error()})
            return
        }

        // ... other error mappings
    }
}
```

---

## Authentication and Authorization

### Required Middleware

The `Required` middleware enforces scope-based authorization:

```go
// Require "applications" scope for all routes
routeGroup.Use(Required("applications"))

// Nested scope for sub-resources
routeGroup.Use(Required("applications.facts"))
```

**How it works:**
1. Extracts bearer token from `Authorization` header
2. Validates token with auth provider
3. Checks if user has required scope for HTTP method
4. Populates `RichContext` with user and granted scopes
5. Returns 401 Unauthorized if auth fails

### Current User

Get the authenticated user in handlers:

```go
m.CreateUser = h.CurrentUser(ctx)
m.UpdateUser = h.CurrentUser(ctx)
```

### Scope Checking

Check for additional scopes in handler logic:

```go
if h.HasScope(ctx, "applications:decrypt") {
    err = h.Decrypt(ctx, m)
    if err != nil {
        _ = ctx.Error(err)
        return
    }
}
```

---

## Associations and Relationships

### Many-to-Many Associations

Use GORM's `Association` API for explicit control:

```go
func (h StakeholderHandler) Create(ctx *gin.Context) {
    r := &Stakeholder{}
    err := h.Bind(ctx, r)
    if err != nil {
        _ = ctx.Error(err)
        return
    }
    m := r.Model()
    m.CreateUser = h.CurrentUser(ctx)

    // Create without associations
    err = h.DB(ctx).Omit(clause.Associations).Create(m).Error
    if err != nil {
        _ = ctx.Error(err)
        return
    }

    // Manage associations explicitly
    err = h.DB(ctx).Model(m).Association("Groups").Replace(m.Groups)
    if err != nil {
        _ = ctx.Error(err)
        return
    }

    err = h.DB(ctx).Model(m).Association("Owns").Replace(m.Owns)
    if err != nil {
        _ = ctx.Error(err)
        return
    }

    r.With(m)
    h.Respond(ctx, http.StatusCreated, r)
}
```

### Association Helper

Use the `Association` helper for cleaner code:

```go
assoc := h.Association(ctx, "FieldName")
err = assoc.Replace(models)
```

### Preloading Associations

Load associations with `preLoad`:

```go
// Load specific associations
db := h.preLoad(h.DB(ctx), "Association1", "Association2")

// Load all associations
db := h.preLoad(h.DB(ctx), clause.Associations)

// Then query
result := db.First(m, id)
```

---

## Sub-Resources

### Nested Resource Routes

Some resources have nested sub-resources:

```go
func (h ApplicationHandler) AddRoutes(e *gin.Engine) {
    // Main resource routes
    routeGroup := e.Group("/")
    routeGroup.Use(Required("applications"), Transaction)
    routeGroup.GET(api.ApplicationsRoute, h.List)
    routeGroup.POST(api.ApplicationsRoute, h.Create)

    // Tags sub-resource with separate scope
    routeGroup = e.Group("/")
    routeGroup.Use(Required("applications"), Transaction)
    routeGroup.GET(api.ApplicationTagsRoute, h.TagList)
    routeGroup.POST(api.ApplicationTagsRoute, h.TagAdd)
    routeGroup.DELETE(api.ApplicationTagRoute, h.TagDelete)

    // Facts sub-resource with different scope
    routeGroup = e.Group("/")
    routeGroup.Use(Required("applications.facts"), Transaction)
    routeGroup.GET(api.ApplicationFactsRoute, h.FactGet)
    routeGroup.POST(api.ApplicationFactsRoute, h.FactCreate)
}
```

### Sub-Resource Handler Methods

```go
// TagList lists tags for an application
// Route: GET /applications/{id}/tags
func (h ApplicationHandler) TagList(ctx *gin.Context) {
    id := h.pk(ctx)
    // ... implementation
}

// TagAdd adds a tag to an application
// Route: POST /applications/{id}/tags
func (h ApplicationHandler) TagAdd(ctx *gin.Context) {
    id := h.pk(ctx)
    // ... implementation
}
```

---

## Pagination and Filtering

### Pagination with Page

Use the `Page` struct for offset/limit pagination:

```go
func (h ResourceHandler) List(ctx *gin.Context) {
    page := Page{}
    page.With(ctx)  // Extracts ?offset=X&limit=Y from query

    var count int64
    db := h.DB(ctx)
    db = db.Model(&model.Resource{})
    db.Count(&count)

    err := h.WithCount(ctx, count)
    if err != nil {
        _ = ctx.Error(err)
        return
    }

    var list []model.Resource
    db = page.Paginated(db)  // Apply offset/limit
    result := db.Find(&list)
    // ... convert and respond
}
```

**The `WithCount` method:**
- Sets `X-Total` header with total count
- Returns error if count exceeds `MaxPage` (500) without pagination
- Returns error if count exceeds `MaxCount` (50000)

### Filtering with Filter Package

Use the filter package for query filtering:

```go
import qf "github.com/konveyor/tackle2-hub/internal/api/filter"

func (h ApplicationHandler) List(ctx *gin.Context) {
    filter, err := qf.New(ctx,
        []qf.Assert{
            {Field: "name", Kind: qf.STRING},
            {Field: "platform.id", Kind: qf.LITERAL},
            {Field: "repository.url", Kind: qf.STRING},
        })
    if err != nil {
        _ = ctx.Error(err)
        return
    }

    // Rename field if needed
    filter = filter.Renamed("platform.id", "PlatformId")

    // Apply filter to query
    db := h.DB(ctx)
    db = filter.ApplyTo(db)

    // ... rest of query
}
```

**Filter syntax in URL:**
```text
?filter=name:MyApp
?filter=name~Java,platform.id:1
```

### Cursor-Based Iteration

For large result sets, use `Cursor` for efficient iteration:

```go
page := Page{}
page.With(ctx)

cursor := Cursor{}
cursor.With(db, page)
defer cursor.Close()

for cursor.Next(&m) {
    if cursor.Error != nil {
        _ = ctx.Error(cursor.Error)
        return
    }
    // Process m
}

count := cursor.Count()
err := h.WithCount(ctx, count)
```

---

## Godoc and Swagger Annotations

### Required Annotations

**Every handler method MUST have godoc and swagger annotations:**

```go
// Get godoc
// @summary Get a resource by ID.
// @description Get a resource by ID.
// @tags resources
// @produce json
// @success 200 {object} api.Resource
// @router /resources/{id} [get]
// @param id path int true "Resource ID"
func (h ResourceHandler) Get(ctx *gin.Context) {
    // Implementation
}
```

### Common Annotation Patterns

**GET (single):**
```go
// @summary Get a resource by ID.
// @description Get a resource by ID.
// @tags resources
// @produce json
// @success 200 {object} api.Resource
// @router /resources/{id} [get]
// @param id path int true "Resource ID"
```

**GET (list):**
```go
// @summary List all resources.
// @description List all resources.
// @tags resources
// @produce json
// @success 200 {object} []api.Resource
// @router /resources [get]
```

**POST:**
```go
// @summary Create a resource.
// @description Create a resource.
// @tags resources
// @accept json
// @produce json
// @success 201 {object} api.Resource
// @router /resources [post]
// @param resource body api.Resource true "Resource data"
```

**PUT:**
```go
// @summary Update a resource.
// @description Update a resource.
// @tags resources
// @accept json
// @success 204
// @router /resources/{id} [put]
// @param id path int true "Resource ID"
// @param resource body api.Resource true "Resource data"
```

**DELETE:**
```go
// @summary Delete a resource.
// @description Delete a resource.
// @tags resources
// @success 204
// @router /resources/{id} [delete]
// @param id path int true "Resource ID"
```

---

## Testing API Handlers

### Test File Organization

- Tests go in `api_test.go` or feature-specific files (e.g., `mapping_test.go`)
- Use Gomega for assertions
- Focus on unit testing helper functions and utilities

### Testing Patterns

```go
package api

import (
    "testing"
    "github.com/onsi/gomega"
    "github.com/gin-gonic/gin"
    "net/http"
)

func TestAccepted(t *testing.T) {
    g := gomega.NewGomegaWithT(t)
    h := BaseHandler{}
    ctx := &gin.Context{
        Request: &http.Request{
            Header: http.Header{},
        },
    }

    ctx.Request.Header[Accept] = []string{"application/json"}
    g.Expect(h.Accepted(ctx, "application/json")).To(gomega.BeTrue())
    g.Expect(h.Accepted(ctx, "text/html")).To(gomega.BeFalse())
}
```

### Integration Testing

- Full API integration tests live in `test/binding/` (outside internal/api)
- Use the client binding to test full request/response cycle
- Test against real database and auth

---

## Summary

### Quick Start

**To add a new endpoint, follow the [complete example](#adding-a-new-endpoint---complete-example) which shows all files you need to create/modify.**

### Files to Create/Modify Checklist

When implementing a new API endpoint:

**Create these files:**
- [ ] `internal/api/{resource}.go` - Handler implementation
- [ ] `internal/api/resource/{resource}.go` - Resource converter

**Modify these files:**
- [ ] `internal/migration/v{N}/model/*.go` - Add model definition
- [ ] `internal/model/pkg.go` - Re-export model type alias
- [ ] `shared/api/*.go` - Add REST API type definition
- [ ] `shared/api/pkg.go` - Add route constants
- [ ] `internal/api/pkg.go` - Register handler in `All()`

### Implementation Checklist

- [ ] Create handler struct embedding `BaseHandler` or `BucketOwner`
- [ ] Implement `AddRoutes()` with proper middleware
- [ ] Implement CRUD methods (Get, List, Create, Update, Delete)
- [ ] Add godoc comment starting with method name
- [ ] Add swagger annotations (@summary, @description, @tags, etc.)
- [ ] Use `h.pk(ctx)` for ID extraction
- [ ] Use `h.Bind(ctx, r)` for request binding
- [ ] Use `h.DB(ctx)` for database access
- [ ] Set `CreateUser`/`UpdateUser` from `h.CurrentUser(ctx)`
- [ ] Use `ctx.Error(err)` for all errors (never handle inline)
- [ ] Use `h.Respond()` or `h.Status()` for responses
- [ ] Use correct HTTP status codes (200, 201, 204)
- [ ] Use `Omit(clause.Associations)` in Update
- [ ] Manage associations explicitly if needed
- [ ] Add handler to `pkg.go:All()`
- [ ] Follow all conventions from main CLAUDE.md

### Common Mistakes to Avoid

- ❌ Handling errors inline instead of using `ctx.Error()`
- ❌ Forgetting to set `CreateUser`/`UpdateUser`
- ❌ Not using `Omit(clause.Associations)` in Update
- ❌ Wrong HTTP status codes (e.g., 200 instead of 201 for Create)
- ❌ Missing swagger annotations
- ❌ Missing `Required()` middleware
- ❌ Not preloading associations before converting to resource
- ❌ Creating new patterns instead of following established ones

### File References

- `base.go` - BaseHandler and common utilities
- `pkg.go` - Handler registration and constants
- `error.go` - Custom error types and ErrorHandler
- `context.go` - RichContext and middleware
- `auth.go` - Required middleware and auth handlers
- `resource/` - Resource conversion patterns
- `filter/` - Query filtering implementation
- `sort/` - Query sorting implementation
- `association/` - Association management utilities

For general Go coding standards, see [../../CLAUDE.md](../../CLAUDE.md)
