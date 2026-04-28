# Auth Package

The auth package provides authentication implementations for the Hub API client.

## Authenticator Interface

All authentication methods implement the `Authenticator` interface:

```go
type Authenticator interface {
    Login() (err error)        // Authenticate/refresh credentials
    Header() (header string)   // Get Authorization header value
}
```

## Implementations

### APIKey

Static API key authentication for backward compatibility.

```go
auth := auth.NewAPIKey("my-api-key")
client := binding.New(hubURL)
client.Client.Use(auth)
```

**Behavior:**
- `Login()` - No-op (API keys don't expire)
- `Header()` - Returns `"Bearer <key>"`

### OIDC

OAuth2/OIDC authentication with device flow and automatic token refresh.

Uses standard scopes: `openid`, `profile`, `email`, `offline_access`

```go
oidcAuth, _ := auth.NewOIDC(hubURL+"/oidc", "konveyor-cli")

// Perform device login
err := oidcAuth.DeviceLogin(context.Background())

// Use with client
client := binding.New(hubURL)
client.Client.Use(oidcAuth)
```

**Behavior:**
- `Login()` - Refreshes access token using refresh token
- `Header()` - Returns `"Bearer <access_token>"`
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

The OIDC implementation supports RFC 8628 Device Authorization Grant:

1. **Initiate:** Client requests device/user codes
2. **Display:** User code and verification URL printed to console
3. **Poll:** Client polls for authorization completion
4. **Store:** Tokens stored in OIDC instance
5. **Auto-refresh:** On 401, refresh token is used automatically

## Design Benefits

✅ **Clean separation** - Client doesn't manage tokens, just calls interface methods  
✅ **Automatic retry** - 401 triggers refresh and retry transparently  
✅ **No expiry tracking** - Server (401) is source of truth, no clock skew issues  
✅ **Extensible** - Easy to add BasicAuth, mTLS, etc.  
✅ **Thread-safe** - OIDC uses mutex for concurrent requests  
✅ **Stateful** - Token lifecycle managed inside authenticator  

## Migration from Old API

**Before (deprecated):**
```go
client := binding.New(hubURL)
client.Client.Use("my-api-key")  // String API key
```

**After:**
```go
client := binding.New(hubURL)
auth := auth.NewAPIKey("my-api-key")
client.Client.Use(auth)  // Authenticator interface
```

## Adding Custom Authenticators

Implement the `Authenticator` interface:

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
