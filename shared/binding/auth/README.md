# Auth Package

The auth package provides authentication implementations for the Hub API client.

## AuthMethod Interface

All authentication methods implement the `AuthMethod` interface:

```go
type AuthMethod interface {
    Login() (err error)        // Authenticate/refresh credentials
    Header() (header string)   // Get Authorization header value
}
```

## Implementations

### Basic

HTTP Basic authentication with username and password.

```go
auth := auth.NewBasic("username", "password")
client := binding.New(hubURL)
client.Client.Use(auth)
```

**Behavior:**
- `Login()` - No-op (credentials don't expire)
- `Header()` - Returns `"Basic <base64(username:password)>"`

**Important:** Basic auth only works for users defined in the hub's user inventory. It will **not** work for users authenticated via external IdP (Keycloak, Azure AD, etc.). For external IdP users, use Bearer with OIDC device flow.

### Bearer

OAuth2/OIDC bearer token authentication with device flow and automatic token refresh.

Uses standard scopes: `openid`, `profile`, `email`, `offline_access`

```go
// Option 1: Device flow (interactive OIDC login)
bearer, _ := auth.NewBearer(hubURL+"/oidc", "cli")
err := bearer.DeviceLogin(context.Background())
client := binding.New(hubURL)
client.Client.Use(bearer)

// Option 2: Explicit token/API key
bearer, _ := auth.NewBearer(hubURL+"/oidc", "cli")
bearer.Use("my-api-key-or-token")
client := binding.New(hubURL)
client.Client.Use(bearer)
```

**Behavior:**
- `Login()` - Refreshes access token using refresh token (if available)
- `Header()` - Returns `"Bearer <access_token>"`
- `Use(token)` - Set explicit bearer token or API key
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

The Bearer authenticator supports RFC 8628 Device Authorization Grant:

1. **Initiate:** Client requests device/user codes
2. **Display:** User code and verification URL printed to console
3. **Poll:** Client polls for authorization completion
4. **Store:** Tokens stored in Bearer instance
5. **Auto-refresh:** On 401, refresh token is used automatically

## Design Benefits

✅ **Clean separation** - Client doesn't manage tokens, just calls interface methods  
✅ **Automatic retry** - 401 triggers refresh and retry transparently  
✅ **No expiry tracking** - Server (401) is source of truth, no clock skew issues  
✅ **Extensible** - Easy to add custom auth methods (mTLS, digest, etc.)  
✅ **Thread-safe** - Bearer uses mutex for concurrent requests  
✅ **Stateful** - Token lifecycle managed inside auth method  
✅ **Multiple methods** - Basic auth, bearer tokens, OIDC device flow  

## Migration from Old API

**Before (deprecated):**
```go
client := binding.New(hubURL)
client.Client.Use("my-api-key")  // String API key
```

**After (API key as bearer token):**
```go
client := binding.New(hubURL)
bearer, _ := auth.NewBearer(hubURL+"/oidc", "cli")
bearer.Use("my-api-key")
client.Client.Use(bearer)  // AuthMethod interface
```

**Or (basic auth):**
```go
client := binding.New(hubURL)
basic := auth.NewBasic("username", "password")
client.Client.Use(basic)  // AuthMethod interface
```

## Adding Custom Auth Methods

Implement the `AuthMethod` interface:

```go
type CustomAuth struct {
    token string
}

func (c *CustomAuth) Login() (err error) {
    // Refresh logic here
    c.token = getNewToken()
    return
}

func (c *CustomAuth) Header() (header string) {
    return "Bearer " + c.token
}
```

Then use it:

```go
client.Client.Use(&CustomAuth{})
```
