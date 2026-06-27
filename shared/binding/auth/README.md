# Auth Package

The auth package provides authentication implementations for the Hub API client.

## AuthMethod Interface

All authentication methods implement the `AuthMethod` interface:

```go
type AuthMethod interface {
    Login() (err error)                  // Authenticate/refresh credentials
    Header() (header string)             // Get Authorization header value
    SetTransport(tp *http.Transport)     // Set HTTP transport for auth operations
}
```

## Implementations

### NoAuth

No authentication (for testing or unauthenticated endpoints).

```go
client := binding.New(hubURL)
// NoAuth is the default - no additional setup needed
```

**Behavior:**
- `Login()` - No-op
- `Header()` - Returns empty string
- `SetTransport()` - No-op

### Basic

HTTP Basic authentication with username and password.

```go
basic := auth.NewBasic("username", "password")
client := binding.New(hubURL)
client.Client.Use(basic)
```

**Behavior:**
- `Login()` - No-op (credentials don't expire)
- `Header()` - Returns `"Basic <base64(username:password)>"`
- `SetTransport()` - No-op (doesn't make HTTP calls)

**Important:** Basic auth only works for users defined in the hub's user inventory. It will **not** work for users authenticated via external IdP (Keycloak, Azure AD, etc.). For external IdP users, use OIDC.

### Bearer

Simple bearer token authentication with a static token or API key.

```go
bearer := auth.NewBearer("my-api-key-or-token")
client := binding.New(hubURL)
client.Client.Use(bearer)
```

**Behavior:**
- `Login()` - No-op (token doesn't auto-refresh)
- `Header()` - Returns `"Bearer <token>"`
- `Token()` - Returns the current token
- `SetTransport()` - No-op (doesn't make HTTP calls)

**Use cases:**
- Static API keys
- Pre-obtained access tokens
- Testing with fixed tokens

### OIDC

OAuth2/OIDC bearer token authentication with device flow and automatic token refresh.

Uses standard scopes: `openid`, `profile`, `email`, `offline_access`

```go
// Option 1: Device flow (interactive OIDC login)
oidc := auth.NewOIDC(hubURL+"/oidc", "cli")
err := oidc.DeviceLogin()
if err != nil {
    // Handle error
}
client := binding.New(hubURL)
client.Client.Use(oidc)

// Option 2: Explicit token (bypasses device flow)
oidc := auth.NewOIDC(hubURL+"/oidc", "cli")
oidc.Use("my-pre-obtained-access-token")
client := binding.New(hubURL)
client.Client.Use(oidc)
```

**Behavior:**
- `Login()` - Refreshes access token using refresh token (falls back to DeviceLogin if no refresh token)
- `Header()` - Returns `"Bearer <access_token>"`
- `Use(token)` - Set explicit access token (bypasses device flow)
- `Token()` - Returns the current access token
- `SetTransport()` - Sets HTTP transport for OIDC requests
- `DeviceLogin()` - Initiates RFC 8628 device authorization flow
- Thread-safe with mutex protection

## How It Works

### Every Request

The HTTP client calls `Header()` on every request:

```go
authHeader := r.auth.Header()  // Cheap - just returns current token
request.Header.Set("Authorization", authHeader)
```

### On 401 Unauthorized

When a request returns 401, the client automatically:
1. Calls `Login()` to refresh credentials
2. Calls `Header()` to get fresh auth header
3. Retries the request once

```go
if response.StatusCode == 401 {
    err := r.auth.Login()      // Refresh token
    if err != nil {
        return  // Can't refresh, give up
    }
    authHeader := r.auth.Header()  // Get fresh header
    request.Header.Set("Authorization", authHeader)
    response := client.Do(request)  // Retry once
}
```

**Note:** 403 Forbidden does NOT trigger retry since it means insufficient permissions, not expired credentials.

## Device Authorization Flow

The OIDC authenticator supports RFC 8628 Device Authorization Grant:

1. **Initiate:** Client requests device/user codes from the OIDC provider
2. **Display:** User code and verification URL printed to console
3. **Poll:** Client polls for authorization completion
4. **Store:** Access and refresh tokens stored in OIDC instance
5. **Auto-refresh:** On 401, refresh token is used automatically via `Login()`

## Design Benefits

✅ **Clean separation** - Client doesn't manage tokens, just calls interface methods  
✅ **Automatic retry** - 401 triggers refresh and retry transparently  
✅ **No expiry tracking** - Server (401) is source of truth, no clock skew issues  
✅ **Extensible** - Easy to add custom auth methods (mTLS, digest, etc.)  
✅ **Thread-safe** - OIDC uses mutex for concurrent requests  
✅ **Stateful** - Token lifecycle managed inside auth method  
✅ **Multiple methods** - NoAuth, Basic, static Bearer tokens, OIDC device flow  

## Choosing an Auth Method

| Method | Use Case | Token Management |
|--------|----------|------------------|
| **NoAuth** | Testing, public endpoints | None |
| **Basic** | Internal hub users only | Static credentials |
| **Bearer** | Static API keys, pre-obtained tokens | Static token |
| **OIDC** | External IdP users, interactive CLI tools | Dynamic with auto-refresh |

**Decision tree:**
- **Testing/no auth needed?** → Use NoAuth (default)
- **Have username/password for internal user?** → Use Basic
- **Have a static API key or token?** → Use Bearer
- **Need interactive login with external IdP?** → Use OIDC with DeviceLogin()
- **Have a refresh token?** → Use OIDC with Use() then Login()

## Migration Examples

### Static API Key

**Before (if there was an old string-based API):**
```go
client := binding.New(hubURL)
client.Client.Use("my-api-key")  // Hypothetical old API
```

**After (using Bearer):**
```go
client := binding.New(hubURL)
bearer := auth.NewBearer("my-api-key")
client.Client.Use(bearer)
```

### Basic Authentication

```go
client := binding.New(hubURL)
basic := auth.NewBasic("username", "password")
client.Client.Use(basic)
```

### OIDC Device Flow

```go
client := binding.New(hubURL)
oidc := auth.NewOIDC(hubURL+"/oidc", "cli")
err := oidc.DeviceLogin()
if err != nil {
    log.Fatal(err)
}
client.Client.Use(oidc)
```

## Token Management

The binding provides methods for managing Personal Access Tokens (PATs):

### Create a Token

```go
client := binding.New(hubURL)
// Authenticate first
pat := &api.PAT{
    Name: "my-token",
}
err := client.Token.Create(pat)
// pat.Token now contains the generated token string
```

### List Tokens

```go
tokens, err := client.Token.List()
for _, token := range tokens {
    fmt.Printf("Token ID: %d, Name: %s\n", token.ID, token.Name)
}
```

### Get a Token by ID

```go
token, err := client.Token.Get(tokenID)
```

### Delete a Token

```go
err := client.Token.Delete(tokenID)
```

### Revoke a Token

Revokes a token and its associated grant (if any). This is more thorough than Delete, as it also removes the underlying OAuth grant.

```go
err := client.Token.Revoke(tokenID)
```

**Revoke vs Delete:**
- **Revoke**: Removes token and associated grant (recommended for cleanup)
- **Delete**: Removes only the token record

## Adding Custom Auth Methods

Implement the `AuthMethod` interface:

```go
type CustomAuth struct {
    token     string
    transport *http.Transport
}

func (c *CustomAuth) Login() (err error) {
    // Refresh logic here - use c.transport if you need to make HTTP calls
    c.token = getNewToken()
    return
}

func (c *CustomAuth) Header() (header string) {
    return "Bearer " + c.token
}

func (c *CustomAuth) SetTransport(tp *http.Transport) {
    c.transport = tp
}
```

Then use it:

```go
client := binding.New(hubURL)
client.Client.Use(&CustomAuth{})
```

**Note:** The client will call `SetTransport()` before `Login()` when handling 401 responses, allowing your auth method to use the same HTTP transport configuration (proxy settings, TLS config, etc.) as the main client.
