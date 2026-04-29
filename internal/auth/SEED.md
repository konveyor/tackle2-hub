# RBAC Seeding System

## Overview

The RBAC seeding system automatically maintains permissions, roles, and users at runtime by:
- **Discovering permissions** from registered API routes
- **Loading roles** from `roles.yaml` with permissions resolved by scope
- **Loading users** from `users.yaml` with roles resolved by name

All seeding is triggered by calling `Domain.Seed()` after route registration completes.

## Architecture

### Scope Registration

During route setup, each `Required("resource")` middleware call registers the resource scope:

```go
func Required(scope string) func(*gin.Context) {
    auth.RegisterScope(scope)  // Adds to registeredScopes map
    return func(ctx *gin.Context) {
        // ... authorization logic
    }
}
```

Example registrations:
```go
Required("applications")        // → "applications"
Required("applications.facts")  // → "applications.facts"
Required("tasks")               // → "tasks"
Required("tasks.bucket")        // → "tasks.bucket"
```

### Seeding Sequence

When `Domain.Seed()` is called, it executes these steps in order:

1. **Permission Seeding**
   - Reads all registered scopes from `registeredScopes`
   - Generates 5 permissions per scope (delete, get, patch, post, put)
   - Reconciles with database: preserves existing IDs, deletes orphans, creates new
   - Builds `permByScope` map: `{"applications:get": 27, "tasks:post": 209, ...}`

2. **Role Seeding**
   - Reads role definitions from `roles.yaml`
   - For each role's resources+verbs, looks up permission IDs from `permByScope`
   - Reconciles with database: preserves existing IDs, deletes orphaned seeded roles, creates/updates
   - Static IDs < 1000 from YAML, user-created roles start at 1001

3. **Role Map Building**
   - Reads all roles from database
   - Builds `roleByName` map: `{"admin": 1, "architect": 2, ...}`

4. **User Seeding**
   - Reads user definitions from `users.yaml`
   - For each user's roles, looks up role IDs from `roleByName`
   - Reconciles with database: preserves existing IDs, deletes orphaned seeded users, creates/updates
   - Hashes passwords, generates UUID subjects
   - Static IDs < 1000 from YAML, user-created users start at 1001

## Permission Generation

For each registered scope, 5 permissions are generated:

**Input:** `["applications", "tasks.bucket"]`

**Output:**
```
applications:delete
applications:get
applications:patch
applications:post
applications:put
tasks.bucket:delete
tasks.bucket:get
tasks.bucket:patch
tasks.bucket:post
tasks.bucket:put
```

**ID Assignment:**
- Existing permissions keep their IDs
- New permissions get next available ID after current max
- Orphaned permissions (routes removed) are deleted

## Role Definition Format

Roles in `roles.yaml`:

```yaml
- id: 1
  role: admin
  resources:
    - name: applications
      verbs: [delete, get, post, put]
    - name: tasks
      verbs: [delete, get, post, put, patch]
```

**Resolution process:**
1. For resource "applications" + verb "get" → look up "applications:get" in `permByScope` → get permission ID
2. Repeat for all resource+verb combinations
3. Associate role with all resolved permission IDs

## User Definition Format

Users in `users.yaml`:

```yaml
- id: 1
  userid: admin
  password: admin
  roles:
    - admin
```

**Resolution process:**
1. For role "admin" → look up "admin" in `roleByName` → get role ID
2. Repeat for all roles
3. Associate user with all resolved role IDs
4. Hash password using `secret.HashPassword()`
5. Generate unique Subject UUID

## ID Preservation Strategy

All seeded resources use a two-tier ID system:

**Seeded resources (ID < 1000):**
- Static IDs from YAML files (roles, users) or sequential (permissions)
- Managed by seeding system
- Deleted if removed from YAML
- IDs preserved across runs

**User-created resources (ID >= 1001):**
- Auto-assigned by database
- Never touched by seeding system
- Safe from deletion
- User has full control

This ensures seeding can safely update predefined resources without affecting user data.

## In-Memory Maps

The seeding system uses two in-memory maps for fast lookups:

**permByScope:** `map[string]uint`
- Built after permission seeding
- Maps scope strings to permission IDs
- Used by role seeding to resolve permissions
- Example: `{"applications:get": 27, "tasks:post": 209}`

**roleByName:** `map[string]uint`
- Built after role seeding
- Maps role names to role IDs
- Used by user seeding to resolve roles
- Example: `{"admin": 1, "architect": 2, "migrator": 3}`

This eliminates N database queries during role and user seeding, making the process very fast.

## Transaction Safety

All database modifications happen within transactions:
- Permission reconciliation: single transaction
- Role reconciliation: single transaction
- User reconciliation: single transaction

If any step fails, changes are rolled back atomically.

## Example Flow

Given these route registrations:
```go
Required("applications")
Required("tasks")
```

**Step 1: Permission Seeding**
```
Generate: applications:{delete,get,patch,post,put}, tasks:{delete,get,patch,post,put}
Create permissions with IDs: 1-10
Build map: {"applications:get": 2, "tasks:post": 9, ...}
```

**Step 2: Role Seeding**
```yaml
- id: 1
  role: admin
  resources:
    - name: applications
      verbs: [get, post]
```
```
Resolve: applications:get → ID 2, applications:post → ID 4
Create role ID 1 with permissions [2, 4]
Build map: {"admin": 1}
```

**Step 3: User Seeding**
```yaml
- id: 1
  userid: admin
  password: admin
  roles:
    - admin
```
```
Resolve: admin → ID 1
Hash password: admin → $2a$10$...
Generate subject: 550e8400-e29b-41d4-a716-446655440000
Create user ID 1 with role [1]
```

## Reconciliation Logic

For each resource type (permissions, roles, users), reconciliation:

1. **Read existing** from database
2. **Read desired** from source (generated or YAML)
3. **Diff** to find what to delete/update/create
4. **Delete** orphaned seeded resources (ID < 1000 not in desired)
5. **Update** existing seeded resources (ID < 1000 in both)
6. **Create** new seeded resources with static IDs from YAML
7. **Ignore** user-created resources (ID >= 1001)

This ensures the database matches the desired state while preserving user data.
