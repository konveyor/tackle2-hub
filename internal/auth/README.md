# Tackle2 Hub Authentication

This document describes the design and architecture of the authentication and authorization system for Tackle2 Hub.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Authentication Methods](#authentication-methods)
- [Subject and Identity Resolution](#subject-and-identity-resolution)
- [OIDC Flows](#oidc-flows)
- [Device Authorization Grant](#device-authorization-grant)
- [IdP Federation](#idp-federation)
- [LDAP Authentication](#ldap-authentication)
- [Storage Architecture](#storage-architecture)
- [Token Types](#token-types)
- [Key Management](#key-management)
- [Session Management](#session-management)
- [Web UI Pages](#web-ui-pages)
- [API Client Integration](#api-client-integration)
- [RBAC Seeding System](#rbac-seeding-system)
- [Role Permissions](#role-permissions)

---

## Architecture Overview

Tackle2 Hub provides a **built-in OIDC (OpenID Connect) provider** that implements OAuth 2.0 and OIDC standards. The authentication system supports multiple flows and can optionally federate to an external identity provider or LDAP server.

```mermaid
graph TB
    subgraph "Tackle2 Hub"
        Provider[OIDC Provider]
        Storage[Storage Layer]
        Cache[In-Memory Cache]
        DAG[Device Auth Handler]
        IDP[IdP Federation Handler]
        LDAP[LDAP Handler]
        KeyMgr[Key Manager]
        
        Provider --> Storage
        Provider --> KeyMgr
        Storage --> Cache
        Storage --> LDAP
        DAG --> Storage
        IDP --> Storage
    end
    
    WebUI[Web UI] --> Provider
    CLI[CLI Tools] --> DAG
    API[API Clients] --> Provider
    ExtIDP[External IdP] -.-> IDP
    LDAPServer[LDAP Server] -.-> LDAP
    
    Storage --> DB[(Database)]
    Cache --> Memory[(Memory)]
```

### Components

| Component | Purpose |
|-----------|---------|
| **OIDC Provider** | Core OAuth 2.0 and OIDC provider implementation |
| **Storage Layer** | Manages authorization requests, grants, and tokens |
| **Cache** | In-memory cache for users, roles, identities, and tokens |
| **Device Auth Handler** | RFC 8628 Device Authorization Grant flow |
| **IdP Federation Handler** | Delegates authentication to external OIDC providers |
| **LDAP Handler** | Authenticates users against LDAP/Active Directory |
| **Key Manager** | RSA key generation, rotation, and JWT signing |

### Deployment Topologies

| Deployment | Auth REST API | OIDC Endpoints | Notes |
|------------|---------------|----------------|-------|
| **Direct Hub Access** | `/auth/*` | `/oidc/*` | Non-OIDC resources (users, roles, etc.) at `/auth`, OIDC endpoints at `/oidc` |
| **Kubernetes/OpenShift (with Route)** | `/hub/auth/*` | `/oidc/*` | `/auth` routes to external IdP (e.g., Keycloak), hub resources at `/hub/auth` |

**Note:** In all deployments, OIDC endpoints are at `/oidc/*`. When deployed in a cluster with a route/ingress, `/auth` typically routes to an external IdP, and hub auth resources (users, roles, etc.) are accessed at `/hub/auth/*`.

---

## Authentication Methods

Tackle2 Hub supports four authentication methods:

### 1. OIDC Tokens (Primary)
OAuth 2.0 access tokens issued by the built-in provider. Used by web UI and API clients for standard authentication flows.

### 2. Basic Authentication
Username/password authentication for local users or LDAP users. Credentials validated against:
- Local users: hashed passwords in database
- LDAP users: authentication against LDAP server

### 3. Personal Access Tokens (PATs)
Long-lived API keys created by users for scripting and automation. Managed via `/auth/token` CRUD endpoints.

### 4. Task API Keys
Short-lived tokens automatically generated for addon task execution. Scoped to task-specific permissions and cleaned up when task completes.

---

## Authentication Staleness

Different authentication methods have different staleness characteristics based on their use cases:

| Authentication Method | Staleness | Environment Variable | Default |
|----------------------|-----------|---------------------|---------|
| **OAuth Access Tokens** | Token lifespan | `OIDC_TOKEN_LIFESPAN` | 5 minutes |
| **Basic Auth (LDAP)** | Identity expiration | `AUTH_BASIC_AUTH_LIFESPAN` | 1 minute |
| **Basic Auth (Local)** | Cache refresh | `AUTH_CACHE_LIFESPAN` | 5 minutes |
| **OAuth Refresh (LDAP)** | Token lifespan | `OIDC_TOKEN_LIFESPAN` | 5 minutes |
| **Personal Access Tokens** | Cache refresh | `AUTH_CACHE_LIFESPAN` | 5 minutes |

### Staleness Behavior

**OAuth Access Tokens:**
- Signed JWTs with embedded scopes
- Valid until expiration (typically 5 minutes)
- Group/role changes don't apply until token refresh
- Designed staleness for performance

**Basic Auth with LDAP:**
- Creates LDAP Identity with expiration based on `BasicAuthLifespan`
- Cached identity reused until expiration
- On expiration: re-authenticate with LDAP, update groups/roles
- Shorter staleness (1 min) for responsive credential/permission changes

**Basic Auth with Local Users:**
- User lookup always refreshes from database (security-critical)
- Password and role changes apply immediately
- Cache used for role→scope resolution only

**Token Refresh with LDAP:**
- Uses Token.Lifespan for Identity expiration
- Re-authenticates with LDAP when identity expires during refresh
- Longer staleness (5 min) aligned with access token validity

**Cache Safety Net:**
- All cached data refreshes from database when `CacheLifespan` exceeded
- Explicit notifications (IdentitySaved, UserSaved, etc.) update cache immediately
- CacheLifespan is worst-case staleness for notification failures

### Lifespan vs Staleness

**Lifespan** = how long a caller controls when creating authentication data:
- `LdapHandler.Authenticate(login, password, lifespan)` sets Identity.Expiration

**Staleness** = resulting delay before changes take effect:
- Basic Auth passes `BasicAuthLifespan` (1 min) → 1 min staleness
- Token refresh passes `TokenLifespan` (5 min) → 5 min staleness

### Environment Variables

Complete list of authentication-related environment variables:

| Variable | Type | Default | Description |
|----------|------|---------|-------------|
| `AUTH_ENABLED` | Boolean | `true` | Enable authentication system |
| `AUTH_REQUIRED` | Boolean | `false` | Enforce authentication for all API requests |
| `AUTH_CACHE_LIFESPAN` | Integer | `5` | Cache refresh interval in minutes |
| `AUTH_BASIC_AUTH_LIFESPAN` | Integer | `60` | Basic auth identity lifespan for LDAP users in seconds |
| `OIDC_ISSUER` | String | `http://localhost:8080` | OIDC issuer URL (hub base URL) |
| `OIDC_TOKEN_LIFESPAN` | Integer | `300` | OAuth access token lifespan in seconds (5 minutes) |
| `OIDC_REFRESH_TOKEN_LIFESPAN` | Integer | `172800` | OAuth refresh token lifespan in seconds (2 days) |
| `OIDC_KEY_ROTATION` | Integer | `90` | RSA signing key rotation interval in days |
| `OIDC_REDIRECT_URI_WEBUI` | String | (issuer URL) | Redirect URI for Web UI OIDC client (defaults to issuer URL) |
| `OIDC_REDIRECT_URI_KAI` | String | `vscode://konveyor.konveyor-ide/auth` | Redirect URI for KAI OIDC client |
| `APIKEY_SECRET` | String | `tackle` | Secret used for API key generation |
| `APIKEY_LIFESPAN` | Integer | `87600` | Personal Access Token lifespan in hours (10 years) |

**Note:** See [shared/settings/README.md](../../shared/settings/README.md) for complete settings documentation.

---

## Subject and Identity Resolution

### Core Concepts

The authentication system uses **Subject** as the central abstraction representing an authenticated entity. A Subject is resolved from one of two sources:

| Source | Description | Use Case |
|--------|-------------|----------|
| **User** | Local hub user with role assignments | Hub-managed authentication |
| **IdpIdentity** | External identity (LDAP or federated OIDC) | LDAP or external IdP authentication |

**Key principle**: A Subject is **mutually exclusive** — it references either a User OR an IdpIdentity, never both.

```mermaid
graph TB
    subgraph "Authentication Sources"
        LocalUser[Local User<br/>username/password]
        LDAPUser[LDAP User<br/>LDAP credentials]
        FedUser[Federated User<br/>external OIDC]
    end
    
    subgraph "Data Layer"
        UserTable[(User Table)]
        IdentityTable[(IdpIdentity Table)]
    end
    
    subgraph "Resolution"
        SubjectCache[Subject Cache]
        SubjectAbstraction[Subject<br/>name, email, roles, scopes]
    end
    
    subgraph "Output"
        JWT[JWT Access Token<br/>with scopes]
    end
    
    LocalUser --> UserTable
    LDAPUser --> IdentityTable
    FedUser --> IdentityTable
    
    UserTable --> SubjectCache
    IdentityTable --> SubjectCache
    
    SubjectCache --> SubjectAbstraction
    SubjectAbstraction --> JWT
```

### Subject Abstraction

A Subject contains the following information:

| Attribute | Description |
|-----------|-------------|
| **Key** | Subject identifier (UUID for User, hash/external subject for IdpIdentity) |
| **Email** | User's email address |
| **Scopes** | Permission scopes (injected into JWT tokens) |
| **Source Reference** | Points to either User or IdpIdentity record (UserId/IdentityId) |

### Subject Derivation

The `Subject` field is the **unique identifier** for an authenticated entity and appears in JWT tokens as the `sub` claim. Different authentication methods derive subjects differently:

| Authentication Method | Subject Format | Example | Derivation |
|----------------------|----------------|---------|------------|
| **Local User** | UUID v4 | `550e8400-e29b-41d4-a716-446655440000` | `uuid.New().String()` |
| **OIDC (External IdP)** | IdP subject | `f8e3a2b1-4c5d-6e7f-8a9b-0c1d2e3f4a5b` | External IdP's `sub` claim (as-is) |
| **LDAP** | HMAC-SHA256 hash | `dsLflKrXRBgaZC3u8XNHJS8UskJ19GM5AWIZ8nBheFA=` | `secret.Hash(login)` |

**Design principles:**
- **Opaque**: Subjects don't reveal PII (username, email, DN) in JWT tokens
- **Stable**: Same user always produces same subject across authentications
- **Unique**: No collisions between different users or authentication methods
- **Globally Unique**: Hub passphrase ensures LDAP hashes are unique to this hub instance

**Why these approaches?**

*Local User (UUID):*
- Random UUID provides natural uniqueness and opacity
- Stable across user's lifetime
- Standard practice for local identity systems

*OIDC (Raw IdP subject):*
- External IdP subject is already opaque per OIDC spec (typically a UUID)
- Must preserve exact value for token refresh and IdP correlation
- Already globally unique within the issuer's namespace
- No hashing needed - would break IdP integration

*LDAP (Hash of login):*
- Raw login identifier would leak PII in JWT tokens
- HMAC with hub's secret passphrase ensures:
  - Opacity: Hash conceals actual login identifier
  - Stability: Login doesn't change when user moves OUs in LDAP
  - Uniqueness: Hub passphrase makes hash unique to this hub instance
  - No collision: External IdP cannot generate matching hash (doesn't know passphrase)
- DN hash was rejected because DN changes break identity (OU reorganizations)

**Collision resistance:**

The different formats provide natural separation:
- UUID vs UUID: ~1 in 5.3 × 10^36 probability (negligible)
- LDAP hash vs OIDC subject: Cryptographically infeasible (HMAC with secret passphrase)
- All three use database uniqueness constraint on `Subject` field

### IdpIdentity Model

The IdpIdentity table stores external authentication mappings for **both** LDAP and federated OIDC users:

| Field | Type | Purpose | LDAP Value | OpenID Value |
|-------|------|---------|------------|--------------|
| **Kind** | String | Discriminator | `"ldap"` | `"openid"` |
| **Issuer** | String | Authentication source | LDAP server URL | OIDC issuer URL |
| **Subject** | String (unique) | Opaque identifier | `secret.Hash(login)` | IdP subject claim |
| **Login** | String | Login identifier | LDAP login | IdP preferred_username |
| **Name** | String | Display name (optional) | From LDAP (optional) | From IdP claims (optional) |
| **Email** | String | Email address | From LDAP | From IdP claims |
| **RefreshToken** | String (encrypted) | Refresh credentials | User password | OAuth refresh token |
| **Expiration** | Timestamp | When identity expires | Based on config | From token expiry |
| **Scopes** | String | Permission scopes | Mapped from LDAP groups | Extracted from access token |

**Why unified model?**
- Single subject resolution path
- Consistent caching strategy
- Unified token refresh mechanism
- Simplified database schema

### Authentication Pipeline

Authentication attempts multiple methods in sequence until one succeeds:

```mermaid
graph LR
    Request[Auth Request] --> Method1[1. Try Local User]
    Method1 -- Not Found --> Method2[2. Try LDAP]
    Method2 -- Not Found --> Method3[3. Try External IdP]
    Method1 -- Success --> Subject[Create Subject]
    Method2 -- Success --> Subject
    Method3 -- Success --> Subject
    Subject --> Token[Issue JWT]
```

**Pipeline behavior:**
- Each method returns: Success, NotFound, or Error
- NotFound → try next method
- Success → create Subject, issue token
- Error → authentication fails

### Cache-Based Resolution

Subjects are resolved from cache for performance:

**Cache structure:**
- Users indexed by subject hash (UUID)
- IdpIdentities indexed by subject hash (login hash for LDAP or external subject for OIDC)
- O(1) lookup by subject identifier
- Auto-refresh when not found or cache expired

**Cache lifecycle:**
1. Subject lookup by identifier
2. If found in cache → return Subject
3. If not found → refresh cache from database
4. Retry lookup
5. If still not found → authentication fails

**Security consideration:**
- User cache **always refreshes** on lookup to prevent stale password usage
- IdpIdentity cache uses standard lifespan to balance security and performance

### Scope Injection

When issuing tokens, permission scopes are injected from the Subject:

```mermaid
sequenceDiagram
    participant Client
    participant TokenEndpoint
    participant Storage
    participant Cache
    participant Database
    
    Client->>TokenEndpoint: Request token
    TokenEndpoint->>Storage: Get subject from request
    Storage->>Cache: FindSubject(hash)
    alt Cache hit
        Cache-->>Storage: Subject with scopes
    else Cache miss
        Cache->>Database: Reload User/IdpIdentity
        Database-->>Cache: Updated data
        Cache-->>Storage: Subject with scopes
    end
    Storage->>TokenEndpoint: Add scopes to token
    TokenEndpoint->>Client: JWT with scopes
```

### Identity Refresh Mechanism

IdpIdentities are automatically refreshed when tokens are refreshed:

| Identity Kind | Refresh Behavior |
|---------------|------------------|
| **Local User** | No refresh needed (roles managed in hub) |
| **LDAP** | Re-authenticate with stored password, fetch fresh group memberships |
| **OpenID** | Use OAuth refresh token to get updated claims from external IdP |

**Why refresh on token refresh?**
- Group memberships may change (LDAP)
- Permission scopes may change (external IdP)
- Ensures current permissions in new access token
- Automatic without user re-authentication

### Subject Resolution Examples

**Local User:**
```
1. User authenticates with username/password
2. Lookup User in database by login
3. Verify password hash
4. Subject.Key = User.Subject (UUID)
5. Subject.Scopes = User.Roles → Permissions.Scope
6. Subject.source = User reference
```

**LDAP User:**
```
1. User authenticates with username/password
2. LDAP server validates credentials, returns DN
3. Fetch group memberships from LDAP
4. Map groups to roles → resolve to scopes
5. Create/update IdpIdentity (Kind=ldap, Subject=hash(login))
6. Subject.Key = hash(login)
7. Subject.Scopes = from IdpIdentity.Scopes
8. Subject.source = IdpIdentity reference
```

**Federated User:**
```
1. OAuth flow completes with external IdP
2. Exchange code for ID token and access token
3. Extract subject, email, scopes from tokens
4. Create/update IdpIdentity (Kind=openid, Subject=IdP subject)
5. Subject.Key = IdP subject claim
6. Subject.Scopes = from IdpIdentity.Scopes
7. Subject.source = IdpIdentity reference
```

### Key Design Decisions

**Why hash LDAP login for subject?**
- Raw login identifier would expose usernames in JWT tokens
- HMAC hash provides stable, opaque identifier
- Same login always produces same subject
- Stable across LDAP OU reorganizations (unlike DN)
- Database uniqueness constraint on subject

**Why store passwords for LDAP users?**
- Token refresh requires re-authentication to get fresh groups
- Alternative: force user to re-login on token expiry
- Password encrypted at ORM level for security
- Enables automatic group membership updates

**Why single IdpIdentity table for LDAP and OpenID?**
- Unified subject resolution
- Consistent caching strategy
- Shared refresh mechanism
- Simpler schema

---

## OIDC Flows

### Authorization Code Flow with PKCE

Standard OAuth 2.0 authorization code flow with PKCE for web applications:

```mermaid
sequenceDiagram
    participant Client
    participant Browser
    participant Hub
    participant Storage
    participant Database

    Client->>Browser: Redirect to /authorize
    Browser->>Hub: GET /authorize?client_id=...&code_challenge=...
    Hub->>Storage: Create auth request (in-memory)
    Storage-->>Hub: authRequestID
    Hub->>Browser: Redirect to /login?authRequestID=...
    
    Browser->>Hub: POST /login (credentials)
    Hub->>Hub: Authenticate (User/LDAP/IdP)
    Hub->>Storage: Update auth request with subject
    Hub->>Browser: Redirect to /authorize/callback
    
    Browser->>Hub: GET /authorize/callback
    Hub->>Storage: Generate authorization code
    Hub->>Browser: Redirect to client with code
    
    Browser->>Client: Return code
    Client->>Hub: POST /token (code, verifier)
    Hub->>Storage: Validate code & PKCE verifier
    Storage->>Database: Create grant & access token
    Storage-->>Hub: Tokens
    Hub-->>Client: access_token, refresh_token
```

**Key characteristics:**
- Auth requests stored **in-memory** (10 minute lifetime)
- PKCE required for security
- Grants and tokens persisted to **database**
- Auth request deleted after code exchange

### Client Credentials Flow

Service-to-service authentication without user context:

```mermaid
sequenceDiagram
    participant Service
    participant Hub
    participant Database

    Service->>Hub: POST /token<br/>grant_type=client_credentials<br/>client_id + client_secret
    Hub->>Hub: Validate credentials
    Hub->>Database: Create access token
    Hub-->>Service: access_token (no refresh token)
```

**Characteristics:**
- No user context
- Direct token issuance
- No refresh token (re-authenticate for new token)
- Scoped to client permissions

### Refresh Token Flow

Extend session without re-authentication:

```mermaid
sequenceDiagram
    participant Client
    participant Hub
    participant Storage
    participant Database

    Client->>Hub: POST /token<br/>grant_type=refresh_token<br/>refresh_token=...
    Hub->>Database: Lookup grant by refresh token hash
    Database-->>Hub: Grant found
    Hub->>Storage: Resolve subject, refresh identity
    Storage->>Storage: Update IdpIdentity if external (LDAP/OpenID)
    Hub->>Database: Create new access token
    Hub-->>Client: access_token, refresh_token
```

**Characteristics:**
- Refresh tokens stored as SHA256 hashes
- New access token issued with current scopes
- IdpIdentities automatically refreshed (LDAP re-auth, OAuth token refresh)
- 30-day default grant lifetime

---

## Device Authorization Grant

Implements **RFC 8628** for devices with limited input capabilities (CLI tools, scripts).

### Overview

Device authorization allows CLI tools and headless clients to authenticate using a user code entered on a separate device (browser).

```mermaid
sequenceDiagram
    participant Device as CLI Device
    participant Hub
    participant Memory as In-Memory Storage
    participant Browser as User Browser
    
    Device->>Hub: POST /device/authorize
    Hub->>Memory: Store device authorization request
    Hub-->>Device: device_code, user_code,<br/>verification_uri
    
    Device->>Device: Display user_code and URI<br/>Start polling /token
    
    Browser->>Hub: Navigate to verification URI
    Hub->>Browser: Login page (if not authenticated)
    Browser->>Hub: Authenticate
    Hub->>Browser: Device verification form
    
    Browser->>Hub: POST user_code
    Hub->>Memory: Lookup by user_code
    Hub->>Memory: Mark authorized with subject
    Hub->>Browser: Success page
    
    Device->>Hub: POST /token (polling)
    Hub->>Memory: Check authorization status
    Memory-->>Hub: Authorized + subject
    Hub->>Database: Create grant & tokens
    Hub-->>Device: access_token, refresh_token
```

### User Code Format

- **Length**: 8 characters
- **Format**: `XXXX-XXXX` (4-4 with dash)
- **Character set**: No vowels (prevents accidental words)
- **Case**: Case-insensitive, normalized to uppercase

### In-Memory Device State

Device authorization requests are stored in memory:

| Field | Purpose |
|-------|---------|
| **device_code** | Unique code for device polling |
| **user_code** | Code displayed to user for entry |
| **client_id** | OAuth client identifier |
| **subject** | Set when user authorizes device |
| **scopes** | Requested permission scopes |
| **expiration** | Default 15 minutes |
| **done** | Authorization completed flag |
| **denied** | User denied authorization flag |

**Storage characteristics:**
- Not persisted to database
- Automatic cleanup of expired requests
- **Hub restart clears all pending authorizations**

### Session-Based Verification

The device verification page requires authenticated users. The hub **acts as an OIDC Relying Party to itself** to manage verification sessions:

```mermaid
graph TB
    subgraph "Hub Dual Role"
        RP[Hub as Relying Party<br/>device-verifier client]
        IdP[Hub as Identity Provider<br/>OIDC provider]
    end
    
    Browser[User Browser] -->|1. Visit /device| RP
    RP -->|2. Check session| Session{Session?}
    Session -->|No| OIDC[3. Initiate OIDC flow]
    OIDC --> IdP
    IdP -->|4. Authenticate| Browser
    Browser -->|5. Callback| RP
    RP -->|6. Set session cookie| Browser
    Session -->|Yes| Form[7. Show code entry form]
    Form --> Browser
```

**Why this design?**
- Device verification page needs authentication
- Can't use standard login form (tied to OAuth auth requests)
- Hub authenticates against itself using OIDC
- Session isolated from main OAuth flows
- Leverages existing authentication (local/LDAP/federated)

### Implementation Details

| Setting | Value | Notes |
|---------|-------|-------|
| **Default lifetime** | 15 minutes | Configurable |
| **Poll interval** | 5 seconds | Recommended by hub |
| **PKCE** | Required | Server-side state, not cookies |
| **Session storage** | Encrypted cookies | AES-GCM with HMAC |

---

## IdP Federation

Optionally delegate authentication to an external OIDC provider (Keycloak, Okta, Azure AD, etc.). Federated users are stored as **IdpIdentity** records with Kind=`"openid"`.

### Architecture

```mermaid
sequenceDiagram
    participant Browser
    participant Hub
    participant ExtIdP as External IdP
    participant Database

    Browser->>Hub: Click "Sign in with IdP"
    Hub->>Browser: Redirect to /idp/login?authRequestID=...
    Hub->>ExtIdP: Redirect to IdP /authorize
    
    Browser->>ExtIdP: Authenticate with IdP
    ExtIdP->>Browser: Redirect to /idp/callback?code=...
    
    Browser->>Hub: GET /idp/callback?code=...
    Hub->>ExtIdP: Exchange code for tokens
    ExtIdP-->>Hub: id_token, access_token, refresh_token
    
    Hub->>Hub: Extract claims (subject, email, scopes)
    Hub->>Database: Create/update IdpIdentity (Kind=openid)
    Hub->>Hub: Complete OAuth authorization request
    Hub->>Browser: Redirect to application with auth code
```

### Configuration

To enable federation to an external OIDC provider, create an `OpenidProvider` Custom Resource:

```yaml
apiVersion: tackle.konveyor.io/v1alpha1
kind: OpenidProvider
metadata:
  name: corporate-sso
  namespace: konveyor-tackle
spec:
  name: "Corporate SSO"
  issuer: "https://idp.example.com/realms/tackle"
  clientId: "tackle-hub"
  clientSecret:
    name: idp-client-secret
    namespace: konveyor-tackle
  redirectURI: "https://hub.example.com/oidc/idp/callback"
  scopes:
    - openid
    - profile
    - email
```

If the external IdP requires a client secret, store it in a Kubernetes Secret:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: idp-client-secret
  namespace: konveyor-tackle
type: Opaque
stringData:
  clientSecret: "<secret>"
```

**Note:** This configures an external provider for federated authentication. The hub's built-in OIDC provider does not require configuration.

| CRD Field | Purpose |
|-----------|---------|
| **name** | Display name for "Sign in with..." button |
| **issuer** | External IdP issuer URL for OIDC discovery |
| **clientId** | OAuth client ID registered with the external IdP |
| **clientSecret** | Reference to Kubernetes Secret containing client secret (optional) |
| **redirectURI** | Callback URL to hub after external IdP authentication |
| **scopes** | OAuth scopes to request from external IdP |

### Identity Lifecycle

**First Authentication:**
1. User clicks "Sign in with [IdP Name]"
2. OAuth flow redirects to external IdP
3. User authenticates with external credentials
4. IdP returns authorization code
5. Hub exchanges code for tokens (ID token, access token, refresh token)
6. Hub extracts claims from tokens
7. **IdpIdentity created** with claims and refresh token
8. Subject created from IdpIdentity
9. Hub OAuth flow completes with hub tokens

**Subsequent Authentications:**
1. Same OAuth flow
2. Claims may have changed at IdP
3. **IdpIdentity updated** (upsert by subject)
4. New hub tokens issued

**Token Refresh:**
1. Client uses hub refresh token
2. Hub looks up IdpIdentity
3. Hub uses **OAuth refresh token** to get fresh tokens from external IdP
4. Claims extracted from new tokens
5. IdpIdentity updated with fresh claims
6. New hub access token issued with current scopes

### Claim Mapping

External IdP claims are mapped to hub identity:

| IdP Claim | IdpIdentity Field | Notes |
|-----------|-------------------|-------|
| **sub** | Subject | Unique identifier from IdP |
| **preferred_username** | Login | Login identifier |
| **email** | Email | Email address |
| **scope** (access token) | Scopes | Permission scopes extracted from access token |

**Scope extraction:**
- Hub extracts `scope` claim from IdP access token
- Scopes are space-separated permission strings
- Stored in IdpIdentity.Scopes field
- Injected into hub-issued access tokens when refreshed

### Security Considerations

- Refresh tokens stored encrypted in database
- PKCE used for authorization code flow
- TLS required for all IdP communication
- Session cookies are HTTP-only, encrypted
- Subject uniqueness enforced (one IdP user → one hub identity)

---

## LDAP Authentication

LDAP authentication allows the hub to authenticate users against an LDAP or Active Directory server. LDAP users are stored as **IdpIdentity** records with Kind=`"ldap"`.

### Architecture

```mermaid
sequenceDiagram
    participant Client
    participant Hub
    participant LDAP as LDAP Server
    participant Database
    
    Client->>Hub: POST /login (login, password)
    Hub->>LDAP: Bind with service account
    LDAP-->>Hub: OK
    
    Hub->>LDAP: Search for user DN
    LDAP-->>Hub: DN=cn=alice,ou=people,dc=example,dc=com
    
    Hub->>LDAP: Bind as user (authenticate)
    LDAP-->>Hub: OK - password valid
    
    Hub->>Hub: Get group memberships
    Hub->>Hub: Map groups → roles → scopes
    
    Hub->>Database: Upsert IdpIdentity<br/>(DN hash, password, scopes)
    Hub->>Client: JWT with scopes
```

### Configuration

To enable federation to an external LDAP or Active Directory server, create an `LdapProvider` Custom Resource:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: ldap-service-account
  namespace: konveyor-tackle
type: Opaque
stringData:
  password: "<service-account-password>"

---
apiVersion: tackle.konveyor.io/v1alpha1
kind: LdapProvider
metadata:
  name: corporate-ldap
  namespace: konveyor-tackle
spec:
  name: "Corporate LDAP"
  kind: "ActiveDirectory"
  url: "ldaps://ldap.example.com:636"
  baseDN: "dc=example,dc=com"
  bindDN: "cn=service,dc=example,dc=com"
  password:
    name: ldap-service-account
    namespace: konveyor-tackle
  userFilter: "(sAMAccountName=%s)"
  groupFilter: "(&(objectClass=group)(member=%s))"
  hasMemberOf: true
  roleMappings:
    - and:
        - "CN=Developers,OU=Groups,DC=example,DC=com"
      roles:
        - architect
    - any:
        - "*Admins*"
      roles:
        - tackle-admin
```

**Note:** This configures an external LDAP/AD server for federated authentication. LDAP users are stored as `IdpIdentity` records, not local hub users.

| CRD Field | Purpose | Default (AD) | Default (LDAP) |
|-----------|---------|--------------|----------------|
| **kind** | Server type ("ACTIVEDIRECTORY", "AD", or blank) | - | - |
| **url** | External LDAP server URL | - | - |
| **baseDN** | Base DN for LDAP searches | - | - |
| **bindDN** | Service account bind DN | - | - |
| **password** | Reference to Secret containing service account password | - | - |
| **userFilter** | User search filter | `(sAMAccountName=%s)` | `(uid=%s)` |
| **groupFilter** | Group search filter | `(&(objectClass=group)(member=%s))` | `(&(objectClass=*)(member=%s))` |
| **hasMemberOf** | Use memberOf attribute for group membership | - | - |
| **roleMappings** | Map LDAP groups to hub roles | - | - |

### LDAP Login as Subject

LDAP users are identified by their login identifier (username), which is hashed for use as the subject:

**Example:**
```
Login from LDAP: alice
Subject hash:    HMAC-SHA256(alice) → dsLflKrXRBgaZC3u8XNHJS8UskJ19GM5AWIZ8nBheFA=
```

**Why hash the login?**
- Raw login identifier would expose usernames in JWT tokens
- HMAC with hub's secret passphrase ensures:
  - Opacity: Hash conceals actual login identifier
  - Stability: Login doesn't change when user moves OUs in LDAP
  - Uniqueness: Hub passphrase makes hash unique to this hub instance
- Same user always has same subject
- Database uniqueness constraint enforced

### Group Membership Strategies

**Option 1: memberOf attribute** (faster, if available):
1. Find user DN
2. Bind as user (authenticate)
3. Read `memberOf` attribute from user entry
4. Parse group DNs from attribute values

**Option 2: Group search** (universal):
1. Find user DN
2. Bind as user (authenticate)
3. Re-bind as service account
4. Search groups where `member=<user DN>`
5. Extract group names from results

**Which to use?**
- Active Directory: Usually has memberOf → set `hasMemberOf: true`
- OpenLDAP: May not have memberOf → set `hasMemberOf: false`

### Role Mapping

LDAP groups are mapped to hub roles using pattern-based rules:

```yaml
roleMappings:
  # AND rule - all patterns must match
  - and:
      - "CN=Developers,*"
      - "CN=Senior,*"
    roles:
      - architect
  
  # OR rule - any pattern must match
  - any:
      - "*Admin*"
      - "*Administrator*"
    roles:
      - tackle-admin
  
  # Combined AND + OR
  - and: ["CN=IT,*"]
    any: ["*PowerUser*", "*Developer*"]
    roles:
      - architect
```

**Rule semantics:**

| Rule Type | Behavior |
|-----------|----------|
| **and** | ALL patterns must match at least one group |
| **any** | At least ONE pattern must match a group |
| **roles** | Roles assigned when rule matches |

**Pattern syntax:**
- Supports glob wildcards (`*`)
- Multiple rules can match (roles accumulated)
- Roles resolved to permission scopes

**Example:**
```
User groups:   ["CN=Developers,OU=Groups,DC=example,DC=com", "CN=Admins,OU=Groups,DC=example,DC=com"]
Rule:          any: ["*Developers*"] → roles: ["architect"]
Rule:          any: ["*Admins*"] → roles: ["tackle-admin"]
Result:        User has roles: architect, tackle-admin
```

### Identity Lifecycle

**First Authentication:**
1. User provides username and password
2. Hub binds to LDAP with service account
3. Search for user by username → get DN
4. Bind as user with password (authenticates)
5. Fetch group memberships (memberOf or search)
6. Map groups to roles via RoleMapper
7. Resolve roles to permission scopes
8. **Create IdpIdentity**:
   - Kind = "ldap"
   - Subject = HMAC-SHA256(login)
   - Login = login
   - RefreshToken = password (encrypted)
   - Scopes = resolved scopes (from role mapping)
   - Expiration = now + lifespan (from caller)
9. Cache notified
10. Subject created from IdpIdentity
11. JWT issued with scopes

**Expiration and Staleness:**

The `Expiration` field controls how long the cached LDAP identity is trusted before re-authentication:

| Flow | Lifespan Parameter | Purpose |
|------|-------------------|---------|
| **Basic Auth** | `BasicAuthLifespan` (1 min) | Fast detection of password/permission changes |
| **Token Refresh** | `TokenLifespan` (5 min) | Aligned with access token validity |
| **OIDC Login** | `TokenLifespan` (5 min) | Matches token refresh behavior |

When identity is found in cache:
- If `Expiration > now`: Use cached scopes (no LDAP contact)
- If expired: Re-authenticate with LDAP, update identity

**Subsequent Authentications:**
1. Check cached IdpIdentity expiration
2. If expired:
   - Same LDAP authentication process
   - Group memberships may have changed
   - Roles/scopes recalculated
   - **IdpIdentity updated** (upsert by subject) with new expiration
3. If not expired: Reuse cached scopes
4. JWT issued with current scopes

**Token Refresh:**
1. Client requests token refresh
2. Hub finds IdpIdentity by subject
3. Check expiration:
   - If expired: **Re-authenticate with LDAP** using stored password, fetch fresh groups
   - If not expired: Reuse cached scopes
4. Issue new access token with current scopes

**Why re-authenticate on refresh?**
- Group memberships may change between token refreshes
- Ensures current permissions without user interaction
- Expiration controls staleness vs LDAP load tradeoff
- Alternative: force re-login on group changes

### Password Storage

LDAP passwords are stored encrypted in the `RefreshToken` field:

| Aspect | Detail |
|--------|--------|
| **Field** | IdpIdentity.RefreshToken |
| **Encryption** | Automatic ORM-level encryption (AES) |
| **Usage** | Re-authentication during token refresh |
| **Security** | Never exposed in API, encrypted at rest |

**Design rationale:**
- Token refresh requires fetching current group memberships
- LDAP server is source of truth for groups
- Re-authentication with stored password enables automatic updates
- Acceptable security trade-off with encryption

### Active Directory vs OpenLDAP

Differences in default configuration:

| Aspect | Active Directory | OpenLDAP |
|--------|------------------|----------|
| **Username attribute** | `sAMAccountName` | `uid` |
| **Group objectClass** | `group` | Generic (`*`) |
| **memberOf support** | Usually present | May not be present |
| **Recommended userFilter** | `(sAMAccountName=%s)` | `(uid=%s)` |
| **Recommended groupFilter** | `(&(objectClass=group)(member=%s))` | `(&(objectClass=*)(member=%s))` |

### Security

**Connection Security:**
- Use `ldaps://` (LDAP over TLS) for encrypted communication
- Validate server certificates in production
- Service account should have minimal privileges (read-only)

**Credential Security:**
- User passwords encrypted at rest
- Passwords transmitted over TLS to LDAP
- Passwords never in JWT tokens
- Passwords never exposed via API

**Authorization:**
- Groups mapped to roles
- Roles resolved to fine-grained scopes
- Scopes embedded in JWT access token
- Group membership refreshed on token refresh

---

## Storage Architecture

The authentication system uses a **hybrid storage model** with in-memory state for transient data and database persistence for durable data.

```mermaid
graph TB
    subgraph "In-Memory Storage"
        AuthReq[Authorization Requests<br/>10 min TTL]
        AuthCode[Auth Codes<br/>maps to requests]
        DevReq[Device Auth Requests<br/>15 min TTL]
        DevCode[User Codes<br/>maps to device codes]
        PKCEState[PKCE State<br/>10 min TTL]
    end
    
    subgraph "Database Storage"
        Users[Users<br/>local credentials]
        Identities[IdpIdentities<br/>LDAP/federated]
        Grants[Grants<br/>refresh tokens]
        Tokens[Tokens<br/>access tokens, PATs]
        Keys[RSA Keys<br/>JWT signing]
    end
    
    Client[Clients] --> AuthReq
    AuthReq --> Grants
    DevReq --> Grants
    
    Users --> Identities
```

### In-Memory Storage

Transient authorization state is stored in memory for performance:

| Data Type | Lifetime | Purpose | Hub Restart Impact |
|-----------|----------|---------|-------------------|
| **Authorization Requests** | 10 minutes | OAuth auth code flow state | All pending flows fail |
| **Device Auth Requests** | 15 minutes | Device code verification state | All pending device flows fail |
| **PKCE State** | 10 minutes | Code verifier for device verification | Verification sessions lost |

**Design rationale:**
- Short-lived state (seconds to minutes)
- High-frequency access during flows
- No need for persistence (user can retry)
- Automatic memory cleanup via expiration

### Database Storage

Long-lived authentication artifacts are persisted:

| Table | Contents | Lifetime | Purpose |
|-------|----------|----------|---------|
| **Users** | Local user credentials, roles | Indefinite | Hub-managed authentication |
| **IdpIdentities** | LDAP/federated user mappings | Until user deleted | External authentication |
| **Grants** | Refresh tokens (hashed) | 30 days (default) | Token refresh |
| **Tokens** | Access tokens, PATs, task keys | Varies by type | API authentication |
| **RsaKeys** | JWT signing keys | Until rotated | Token signing |

**Design rationale:**
- Long-lived data (hours to days to indefinite)
- Must survive hub restarts
- Audit trail for tokens
- Enables token revocation

### Cache Architecture

In-memory cache reduces database load:

| Cached Data | Index | Refresh Strategy |
|-------------|-------|------------------|
| **Users** | By subject, by login | Always on lookup (security) |
| **IdpIdentities** | By subject | Item expiration + cache lifespan |
| **Roles** | By ID, by name | Standard lifespan |
| **Tokens (PATs)** | By digest | Standard lifespan |

**Cache lifespan:** Configurable via `AUTH_CACHE_LIFESPAN` (default 5 min)

**Cache invalidation:**
- Auto-refresh when data not found
- Auto-refresh when lifespan exceeded
- Explicit notification on data changes (UserSaved, IdentitySaved, etc.)

**Why always refresh Users?**
- Password changes must take effect immediately
- Role changes must apply on next authentication
- Security-critical data requires fresh reads

**Why item expiration for IdpIdentities?**
- LDAP identities have individual `Expiration` timestamps
- Expiration based on authentication flow (Basic Auth vs Token Refresh)
- Allows different staleness for different use cases
- Cache lifespan is safety net for notification failures

---

## Token Types

The hub issues several types of tokens for different use cases:

### JWT Access Tokens

Short-lived tokens for API authentication:

| Attribute | Value |
|-----------|-------|
| **Format** | JSON Web Token (JWT) |
| **Signing** | RSA256 with hub private key |
| **Lifetime** | 5 minutes (default, configurable via `OIDC_TOKEN_LIFESPAN`) |
| **Claims** | `sub`, `scope`, `exp`, `iss`, `aud` |
| **Usage** | Bearer token in `Authorization` header |

**Standard claims:**
- `sub`: Subject identifier (UUID for User, hash/external for IdpIdentity)
- `scope`: Space-separated permission scopes
- `exp`: Expiration timestamp
- `iss`: Issuer URL (hub OIDC provider)
- `aud`: Audience (client ID)

### Refresh Tokens

Long-lived opaque tokens for obtaining new access tokens:

| Attribute | Value |
|-----------|-------|
| **Format** | Opaque string (random) |
| **Storage** | SHA256 hash in database |
| **Lifetime** | 30 days (default, configurable) |
| **Rotation** | Optional (new refresh token on use) |
| **Revocation** | Delete from database |

**Security:**
- Stored as hash (prevents token leak from DB compromise)
- Bound to specific grant (client + user)
- Can be revoked independently of access tokens

### Personal Access Tokens (PATs)

User-created API keys for automation:

| Attribute | Value |
|-----------|-------|
| **Format** | Opaque string (random) |
| **Storage** | SHA256 hash in database |
| **Lifetime** | Configurable (default 24 hours, max set by admin) |
| **Scopes** | Inherits creating user's permissions |
| **Management** | CRUD via `/auth/token` endpoints |

**Use cases:**
- CLI scripting
- CI/CD pipelines
- Integration testing
- Automation tools

### Task API Keys

Automatically generated tokens for addon execution:

| Attribute | Value |
|-----------|-------|
| **Format** | Opaque string (random) |
| **Lifetime** | Tied to task execution |
| **Scopes** | Task-specific permissions (restricted) |
| **Cleanup** | Automatic on task completion |

**Scope restrictions:**
- Read-only for most resources
- Write access to task-specific endpoints
- No admin capabilities
- Defined in `AddonScopes` constant

---

## Key Management

The hub uses RSA key pairs for JWT signing.

### RSA Key Generation

**On first startup:**
1. Generate 2048-bit RSA key pair
2. Encode private key as PKCS#1 PEM
3. Encrypt private key using hub encryption key
4. Store in database `RsaKey` table
5. Assign sequential key ID

**Key structure:**

| Field | Purpose |
|-------|---------|
| **ID** | Key identifier (kid in JWT header) |
| **PEM** | Encrypted private key (PKCS#1 PEM format) |
| **CreateTime** | Key generation timestamp |

### JWKS Endpoint

Public keys published at `/.well-known/jwks.json`:

**Response format:**
```json
{
  "keys": [
    {
      "kty": "RSA",
      "use": "sig",
      "kid": "1",
      "n": "<modulus>",
      "e": "AQAB"
    }
  ]
}
```

**Usage:**
- Clients verify JWT signatures using public keys
- Standard OIDC discovery mechanism
- Supports key rotation (multiple keys in set)

### Key Rotation

**Current implementation:**
- Manual rotation (admin generates new key)
- Old keys retained for verification of existing tokens
- Future enhancement: automatic rotation on schedule

**Rotation process:**
1. Generate new RSA key
2. Add to database with new ID
3. New tokens signed with new key
4. Old keys still published in JWKS
5. Old keys eventually deleted after all tokens expire

---

## Session Management

### Device Verification Sessions

The device authorization flow uses **cookie-based sessions** for user authentication.

**Why sessions for device verification?**
- Verification page requires authenticated user
- Standard OAuth flow not applicable (no client context)
- Hub acts as OIDC Relying Party to itself
- Session isolates verification from OAuth flows

### Cookie Security

Session cookies use multiple layers of security:

| Security Feature | Implementation |
|------------------|----------------|
| **Encryption** | AES-GCM encryption of cookie value |
| **Authentication** | HMAC signature to prevent tampering |
| **HTTP-only** | Not accessible via JavaScript (XSS protection) |
| **SameSite** | Lax mode (CSRF protection) |
| **Secure flag** | HTTPS only (production) |

### Key Derivation

Cookie encryption keys are derived from the hub's API key secret:

```
Hash Key (HMAC):      SHA256(secret + "-hash")
Encryption Key (AES): SHA256(secret + "-encrypt")
```

**Benefits:**
- No additional configuration needed
- Consistent keys across hub instances
- Keys rotate when hub secret changes
- Cryptographically strong derivation

### Session Contents

Session cookie stores only the OIDC subject:

**Cookie name:** `oidc_subject`  
**Cookie value:** Authenticated user's subject identifier (encrypted)

**Usage:**
1. User authenticates via hub OIDC flow
2. Subject stored in encrypted session cookie
3. Verification page reads cookie to get subject
4. Device authorization updated with subject
5. Session cookie cleared after verification

### Session Lifetime

- **Lifetime:** Duration of OIDC callback flow (minutes)
- **Persistence:** Not persisted to database
- **Hub restart impact:** All verification sessions lost (acceptable - user can re-authenticate)

---

## Web UI Pages

All authentication pages use consistent, modern styling for a cohesive user experience.

### Login Page (`/oidc/login`)

**Purpose:** Authenticate users for OAuth authorization flows

**Features:**
- Username and password fields
- Optional "Sign in with [IdP]" button (when federation enabled)
- Mobile responsive design
- Form validation

### Device Authorization Page (`/oidc/device`)

**Purpose:** Verify device authorization requests via user code entry

**Features:**
- User code input (8 characters, `XXXX-XXXX` format)
- Monospace input with auto-uppercase
- Automatic redirect to login if not authenticated
- Code validation and completion

### Authorization Success Page

**Purpose:** Confirm successful device authorization

**Features:**
- Checkmark icon
- "Authorization Complete" message
- Instructions to return to device
- No user interaction required

### Design System

All pages share a consistent design:

| Element | Style |
|---------|-------|
| **Background** | Purple gradient (`#667eea` to `#764ba2`) |
| **Container** | White card with shadow and rounded corners |
| **Typography** | System font stack, h1=20px, body=13px |
| **Inputs** | 2px border, rounded corners, focus color `#667eea` |
| **Buttons** | Gradient background, hover animations (transform, shadow) |
| **Layout** | Centered, max-width 400px, responsive padding |

**Accessibility:**
- Semantic HTML
- Proper label associations
- Focus indicators
- Keyboard navigation

---

## API Client Integration

### Bearer Authentication

The `shared/binding/auth` package provides a Bearer authenticator for API clients:

**Features:**
- Device authorization flow
- Automatic token refresh
- Token storage and reuse
- Bearer header injection

**Example usage:**
```bash
# Using hub binding library
bearer := auth.NewBearer(issuerURL, "cli")
bearer.DeviceLogin(context.Background())

# Bearer automatically includes token in requests
client := binding.New(hubURL)
client.Client.Use(bearer)
applications, err := client.Application.List()
```

### CLI Tools

**Example CLI authentication flows:**

| Flow | Command Example | Use Case |
|------|-----------------|----------|
| **Device flow** | `./login -r https://tackle.example.com` | Interactive login |
| **Basic auth** | `./login -login admin -password pass` | Local users |
| **Bearer token** | `./login -b <token>` | Reuse existing token |
| **PAT** | Create via API after device login | Scripting, CI/CD |

**Hub URL variations:**
```bash
# Route deployment (OpenShift/Kubernetes)
./login -r https://tackle.example.com

# Direct hub access
./login -u http://localhost:8080

# Override issuer (rare)
./login -u http://localhost:8080 -i http://localhost:8080/oidc
```

---

## RBAC Seeding System

### Overview

The RBAC system automatically maintains permissions, roles, and users at runtime by:

1. **Discovering permissions** from registered API route scopes
2. **Loading roles** from `roles.yaml` with permission references
3. **Loading users** from `users.yaml` with role references

**Trigger:** `Domain.Seed()` called after route registration completes

### Permission Generation

For each registered scope, 5 permissions are generated:

**Input:** Scope `"applications"`

**Output:** 5 permissions
- `applications:delete`
- `applications:get`
- `applications:patch`
- `applications:post`
- `applications:put`

**HTTP verb mapping:**
- `get` → Read
- `post` → Create
- `put` / `patch` → Update
- `delete` → Delete

### Role Definition Format

Roles defined in `seed/roles.yaml`:

```yaml
- id: 1
  role: admin
  resources:
    - name: applications
      verbs: [delete, get, post, put]
    - name: tasks
      verbs: [delete, get, post, put, patch]
```

**Resolution:**
1. For each resource + verb → lookup permission scope
2. Example: `applications` + `get` → find permission `applications:get`
3. Associate all resolved permissions with role

### User Definition Format

Users defined in `seed/users.yaml`:

```yaml
- id: 1
  login: admin
  password: admin
  roles:
    - admin
```

**Resolution:**
1. For each role name → lookup role by name
2. Associate all resolved roles with user
3. Hash password using bcrypt
4. Generate UUID subject

### ID Preservation

Two-tier ID system ensures safety:

| ID Range | Source | Managed By | Deletable? |
|----------|--------|------------|------------|
| **< 1000** | YAML files | Seeding system | Yes (if removed from YAML) |
| **≥ 1001** | User-created | Hub users | No (never touched by seeding) |

**Benefits:**
- Seeding can safely update predefined resources
- User-created resources never affected
- ID stability across hub restarts
- Clear boundary between seeded and user data

### Reconciliation Logic

For each resource type (permissions, roles, users):

1. **Read** existing from database
2. **Read** desired from source (generated or YAML)
3. **Diff** to find what to delete/update/create
4. **Delete** orphaned seeded resources (ID < 1000, not in desired)
5. **Update** existing seeded resources (ID < 1000, in both)
6. **Create** new seeded resources with static IDs
7. **Ignore** user-created resources (ID ≥ 1001)

**Transaction safety:**
- All changes in single database transaction
- Rollback on any error
- Atomic updates ensure consistency

---

## Role Permissions

This section documents the permissions granted to each predefined role.

**Verb mapping (CRUD):**
- `get` → Read
- `post` → Create
- `put` / `patch` → Update
- `delete` → Delete

### 🛡 Role: tackle-admin

Full administrative access to nearly all resources.

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| addons | ✅ | ✅ | ✅ | ✅ |
| adoptionplans | ✅ | ❌ | ❌ | ❌ |
| analyses | ✅ | ✅ | ✅ | ✅ |
| applications | ✅ | ✅ | ✅ | ✅ |
| applications.analyses | ✅ | ✅ | ✅ | ✅ |
| applications.assessments | ✅ | ✅ | ❌ | ❌ |
| applications.bucket | ✅ | ✅ | ✅ | ✅ |
| applications.facts | ✅ | ✅ | ✅ | ✅ |
| applications.manifests | ✅ | ✅ | ❌ | ❌ |
| applications.stakeholders | ❌ | ❌ | ✅ | ❌ |
| applications.tags | ✅ | ✅ | ✅ | ✅ |
| archetypes | ✅ | ✅ | ✅ | ✅ |
| archetypes.assessments | ✅ | ✅ | ❌ | ❌ |
| assessments | ✅ | ✅ | ✅ | ✅ |
| buckets | ✅ | ✅ | ✅ | ✅ |
| businessservices | ✅ | ✅ | ✅ | ✅ |
| cache | ❌ | ✅ | ❌ | ✅ |
| dependencies | ✅ | ✅ | ✅ | ✅ |
| files | ✅ | ✅ | ✅ | ✅ |
| generators | ✅ | ✅ | ✅ | ✅ |
| identities | ✅ | ✅ | ✅ | ✅ |
| imports | ✅ | ✅ | ✅ | ✅ |
| jobfunctions | ✅ | ✅ | ✅ | ✅ |
| kai | ✅ | ✅ | ❌ | ❌ |
| manifests | ✅ | ✅ | ✅ | ✅ |
| migrationwaves | ✅ | ✅ | ✅ | ✅ |
| platforms | ✅ | ✅ | ✅ | ✅ |
| proxies | ✅ | ✅ | ✅ | ✅ |
| questionnaires | ✅ | ✅ | ✅ | ✅ |
| reviews | ✅ | ✅ | ✅ | ✅ |
| rulesets | ✅ | ✅ | ✅ | ✅ |
| schemas | ✅ | ✅ | ✅ | ✅ |
| settings | ✅ | ✅ | ✅ | ✅ |
| stakeholdergroups | ✅ | ✅ | ✅ | ✅ |
| stakeholders | ✅ | ✅ | ✅ | ✅ |
| tagcategories | ✅ | ✅ | ✅ | ✅ |
| tags | ✅ | ✅ | ✅ | ✅ |
| targets | ✅ | ✅ | ✅ | ✅ |
| tasks | ✅ | ✅ | ✅ | ✅ |
| tasks.bucket | ✅ | ✅ | ✅ | ✅ |
| tickets | ✅ | ✅ | ✅ | ✅ |
| trackers | ✅ | ✅ | ✅ | ✅ |

### 🛠 Role: architect

Broad create/update/delete rights, restricted on sensitive resources (identities, proxies, settings, trackers).

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| addons | ✅ | ✅ | ✅ | ✅ |
| adoptionplans | ✅ | ❌ | ❌ | ❌ |
| analyses | ✅ | ✅ | ✅ | ✅ |
| applications | ✅ | ✅ | ✅ | ✅ |
| applications.analyses | ✅ | ✅ | ✅ | ✅ |
| applications.assessments | ✅ | ✅ | ❌ | ❌ |
| applications.bucket | ✅ | ✅ | ✅ | ✅ |
| applications.facts | ✅ | ✅ | ✅ | ✅ |
| applications.manifests | ✅ | ✅ | ❌ | ❌ |
| applications.stakeholders | ❌ | ❌ | ✅ | ❌ |
| applications.tags | ✅ | ✅ | ✅ | ✅ |
| archetypes | ✅ | ✅ | ✅ | ✅ |
| archetypes.assessments | ✅ | ✅ | ❌ | ❌ |
| assessments | ✅ | ✅ | ✅ | ✅ |
| buckets | ✅ | ✅ | ✅ | ✅ |
| businessservices | ✅ | ✅ | ✅ | ✅ |
| cache | ❌ | ✅ | ❌ | ❌ |
| dependencies | ✅ | ✅ | ✅ | ✅ |
| files | ✅ | ✅ | ✅ | ✅ |
| generators | ✅ | ✅ | ✅ | ✅ |
| identities | ❌ | ✅ | ❌ | ❌ |
| imports | ✅ | ✅ | ✅ | ✅ |
| jobfunctions | ✅ | ✅ | ✅ | ✅ |
| kai | ✅ | ✅ | ❌ | ❌ |
| manifests | ✅ | ✅ | ✅ | ✅ |
| migrationwaves | ✅ | ✅ | ✅ | ✅ |
| platforms | ✅ | ✅ | ✅ | ✅ |
| proxies | ❌ | ✅ | ❌ | ❌ |
| questionnaires | ❌ | ✅ | ❌ | ❌ |
| reviews | ✅ | ✅ | ✅ | ✅ |
| rulesets | ✅ | ✅ | ✅ | ✅ |
| schemas | ❌ | ✅ | ❌ | ❌ |
| settings | ❌ | ✅ | ❌ | ❌ |
| stakeholdergroups | ✅ | ✅ | ✅ | ✅ |
| stakeholders | ✅ | ✅ | ✅ | ✅ |
| tagcategories | ✅ | ✅ | ✅ | ✅ |
| tags | ✅ | ✅ | ✅ | ✅ |
| targets | ✅ | ✅ | ✅ | ✅ |
| tasks | ✅ | ✅ | ✅ | ✅ |
| tasks.bucket | ✅ | ✅ | ✅ | ✅ |
| tickets | ✅ | ✅ | ✅ | ✅ |
| trackers | ❌ | ✅ | ❌ | ❌ |

### 🚚 Role: migrator

Mostly read-only, except full management of dependencies and tasks.

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| addons | ❌ | ✅ | ❌ | ❌ |
| adoptionplans | ✅ | ❌ | ❌ | ❌ |
| analyses | ❌ | ✅ | ❌ | ❌ |
| applications | ❌ | ✅ | ❌ | ❌ |
| applications.analyses | ❌ | ✅ | ❌ | ❌ |
| applications.assessments | ❌ | ✅ | ❌ | ❌ |
| applications.bucket | ❌ | ✅ | ❌ | ❌ |
| applications.facts | ❌ | ✅ | ❌ | ❌ |
| applications.manifests | ❌ | ✅ | ❌ | ❌ |
| applications.tags | ❌ | ✅ | ❌ | ❌ |
| archetypes | ❌ | ✅ | ❌ | ❌ |
| archetypes.assessments | ❌ | ✅ | ❌ | ❌ |
| assessments | ❌ | ✅ | ❌ | ❌ |
| buckets | ❌ | ✅ | ❌ | ❌ |
| businessservices | ❌ | ✅ | ❌ | ❌ |
| cache | ❌ | ✅ | ❌ | ❌ |
| dependencies | ✅ | ✅ | ✅ | ✅ |
| files | ❌ | ✅ | ❌ | ❌ |
| generators | ❌ | ✅ | ❌ | ❌ |
| identities | ❌ | ✅ | ❌ | ❌ |
| imports | ❌ | ✅ | ❌ | ❌ |
| jobfunctions | ❌ | ✅ | ❌ | ❌ |
| kai | ✅ | ✅ | ❌ | ❌ |
| manifests | ❌ | ✅ | ❌ | ❌ |
| migrationwaves | ❌ | ✅ | ❌ | ❌ |
| platforms | ❌ | ✅ | ❌ | ❌ |
| proxies | ❌ | ✅ | ❌ | ❌ |
| questionnaires | ❌ | ✅ | ❌ | ❌ |
| reviews | ❌ | ✅ | ❌ | ❌ |
| rulesets | ❌ | ✅ | ❌ | ❌ |
| schemas | ❌ | ✅ | ❌ | ❌ |
| settings | ❌ | ✅ | ❌ | ❌ |
| stakeholdergroups | ❌ | ✅ | ❌ | ❌ |
| stakeholders | ❌ | ✅ | ❌ | ❌ |
| tagcategories | ❌ | ✅ | ❌ | ❌ |
| tags | ❌ | ✅ | ❌ | ❌ |
| targets | ❌ | ✅ | ❌ | ❌ |
| tasks | ✅ | ✅ | ✅ | ✅ |
| tasks.bucket | ✅ | ✅ | ✅ | ✅ |
| tickets | ❌ | ✅ | ❌ | ❌ |
| trackers | ❌ | ✅ | ❌ | ❌ |

### 📋 Role: project-manager

Read-only for most resources, can update application stakeholders and fully manage migration waves.

| Resource | Create | Read | Update | Delete |
|----------|--------|------|--------|--------|
| addons | ❌ | ✅ | ❌ | ❌ |
| adoptionplans | ✅ | ❌ | ❌ | ❌ |
| applications | ❌ | ✅ | ❌ | ❌ |
| applications.analyses | ❌ | ✅ | ❌ | ❌ |
| applications.assessments | ❌ | ✅ | ❌ | ❌ |
| applications.bucket | ❌ | ✅ | ❌ | ❌ |
| applications.facts | ❌ | ✅ | ❌ | ❌ |
| applications.manifests | ❌ | ✅ | ❌ | ❌ |
| applications.stakeholders | ❌ | ❌ | ✅ | ❌ |
| applications.tags | ❌ | ✅ | ❌ | ❌ |
| archetypes | ❌ | ✅ | ❌ | ❌ |
| archetypes.assessments | ❌ | ✅ | ❌ | ❌ |
| assessments | ❌ | ✅ | ❌ | ❌ |
| buckets | ❌ | ✅ | ❌ | ❌ |
| businessservices | ❌ | ✅ | ❌ | ❌ |
| cache | ❌ | ✅ | ❌ | ❌ |
| dependencies | ❌ | ✅ | ❌ | ❌ |
| generators | ❌ | ✅ | ❌ | ❌ |
| identities | ❌ | ✅ | ❌ | ❌ |
| imports | ❌ | ✅ | ❌ | ❌ |
| jobfunctions | ❌ | ✅ | ❌ | ❌ |
| kai | ✅ | ✅ | ❌ | ❌ |
| migrationwaves | ✅ | ✅ | ✅ | ✅ |
| platforms | ❌ | ✅ | ❌ | ❌ |
| proxies | ❌ | ✅ | ❌ | ❌ |
| questionnaires | ❌ | ✅ | ❌ | ❌ |
| reviews | ❌ | ✅ | ❌ | ❌ |
| rulesets | ❌ | ✅ | ❌ | ❌ |
| schemas | ❌ | ✅ | ❌ | ❌ |
| settings | ❌ | ✅ | ❌ | ❌ |
| stakeholdergroups | ❌ | ✅ | ❌ | ❌ |
| stakeholders | ❌ | ✅ | ❌ | ❌ |
| tagcategories | ❌ | ✅ | ❌ | ❌ |
| tags | ❌ | ✅ | ❌ | ❌ |
| targets | ❌ | ✅ | ❌ | ❌ |
| tasks | ❌ | ✅ | ❌ | ❌ |
| tasks.bucket | ❌ | ✅ | ❌ | ❌ |
| tickets | ❌ | ✅ | ❌ | ❌ |
| trackers | ❌ | ✅ | ❌ | ❌ |
