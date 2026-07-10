# Authentication Frontend

The `internal/frontend/auth/` package contains the login page handler and frontend content.
The frontend is a standalone React + PatternFly 6 project that provides the OIDC login
and device authorization pages. The assets are embedded into the Go binary at compile time.

## Directory Structure

```
internal/frontend/auth/
  handler.go            # Go handler with embedded assets and Render function
  content/              # Frontend project (React + PatternFly 6)
    package.json        # npm project (ESM, @tackle-hub/login-page)
    tsconfig.json       # TypeScript config targeting ES2020 / bundler resolution
    rspack.config.ts    # Rspack 2 build config with branding + asset serving
    branding/           # Default branding (swappable at container build time)
      strings.json      # Branding strings: app title, page titles, image paths
      logo.svg          # Default brand logo
    src/
      index.html        # HTML template source (contains Go template actions)
      index.tsx         # Entry point; reads window.__LOGIN_CONFIG__ and routes to a page
      types.ts          # LoginConfig interface (runtime config injected by hub)
      branding.ts       # Typed access to build-time branding via __BRANDING_STRINGS__
      UserLoginPage.tsx # PF6 login form (username/password + optional federated IdP)
      DeviceVerifyPage.tsx # PF6 device code entry form
      DeviceSuccessPage.tsx # PF6 device authorization success page
    dist/               # Build output (gitignored; embedded via go:embed)
```

> The rspack `output.path` is `internal/frontend/auth/content/dist/`. The Go binary embeds
> these assets at compile time via `go:embed` in `internal/frontend/auth/handler.go`.

## Building the Frontend

```bash
cd internal/frontend/auth/content
npm install
npm run build          # outputs to internal/frontend/auth/content/dist/
```

The hub embeds these assets at compile time via `go:embed` in `handler.go`. You must build the frontend before building the Go binary:

```bash
make frontend  # Builds all frontends including auth
make hub       # Builds Go binary with embedded assets
```

Or manually:
```bash
cd internal/frontend/auth/content && npm run build
cd ../../../.. && go build cmd/main.go
```

## Custom Branding

Branding strings are baked into the bundle at build time by rspack's `DefinePlugin`.
Custom branding is specified via Docker build arg and copied over the default branding
before the frontend build.

```bash
# Default branding (uses internal/frontend/auth/content/branding/)
docker build -t tackle2-hub .

# Custom branding (override with your branding directory)
docker build --build-arg BRANDING=my-custom-branding -t tackle2-hub .
```

### `branding/strings.json` Shape

```json
{
  "application": { "title": "...", "name": "...", "description": "..." },
  "loginPage":   { "title": "...", "subtitle": "..." },
  "devicePage":  { "title": "...", "subtitle": "...", "successTitle": "...", "successMessage": "..." },
  "images":      { "brand": "branding/logo.svg", "background": "" }
}
```

Image assets (SVGs, PNGs) in the branding directory are copied to `dist/branding/`
by `CopyRspackPlugin`. Reference them from `strings.json` using the path
`branding/<filename>` (relative to the served root).

### Dockerfile Branding

See `Dockerfile` for details. The build copies custom branding before running `npm run build`:

```dockerfile
FROM registry.access.redhat.com/ubi10/nodejs-22:latest as login-page
ARG BRANDING=internal/frontend/auth/content/branding
COPY --chown=1001:0 internal/frontend/auth/content/ .
COPY --chown=1001:0 ${BRANDING}/ branding/
RUN npm ci && npm run build
```

## Runtime Configuration Injection

The rspack build outputs `dist/index.html.tmpl` — a Go `text/template` file.
The source `src/index.html` contains Go template actions (e.g. `{{ . }}`)
which rspack passes through unchanged (HTML minification is disabled for this
reason).

On each request, the handler's `Render()` method reads the embedded template, parses it
with Go's `text/template` package, and executes it with the JSON-serialized config.

### `LoginConfig` Interface

TypeScript definition in `content/src/types.ts`, mirrored as `Request` struct in Go:

```typescript
interface LoginConfig {
  page: "login" | "device-verify" | "device-success";
  formAction?: string;       // POST URL for login form
  errorMessage?: string;     // Shown on authentication failure
  federatedIdp?: {
    name: string;
    loginUrl: string;        // Redirect URL for external IdP button
  };
  deviceFormAction?: string; // POST URL for device code form
}
```

### Go API

```go
import "github.com/konveyor/tackle2-hub/internal/frontend/auth"

req := auth.Request{
    Page:         auth.Login,
    FormAction:   "/oidc/login?authRequestId=...",
    ErrorMessage: "Invalid username or password",
}
err := auth.Handler{}.Render(w, req)
```

## Static Assets

The compiled JS/CSS/font assets are embedded in the Go binary and served at
`/frontend/auth/*` (or `/hub/frontend/auth/*` when behind a proxy with base path).

The rspack `publicPath` is set dynamically based on the `HUB_BASE_PATH` environment
variable during build to handle different deployment scenarios.

## Adding a New Page Type

1. Add a new value to `LoginConfig.page` in `content/src/types.ts`.
2. Create `content/src/NewPage.tsx` using PatternFly 6 components.
3. Add a `case` for the new page type in `content/src/index.tsx`.
4. Add the corresponding `page` constant to `handler.go` and use in `Request`.
5. Re-run `npm run build` in `content/` and rebuild the Go binary.

## Key Technology Choices

| Concern | Choice |
|---|---|
| UI framework | React 18 + PatternFly 6 (`@patternfly/react-core`) |
| Bundler | Rspack 2 with `builtin:swc-loader` |
| Branding injection | `DefinePlugin` (`__BRANDING_STRINGS__`) + Docker build arg |
| Asset copy | `CopyRspackPlugin` |
| HTML template | `HtmlRspackPlugin` → outputs `index.html.tmpl` |
| Config injection | Go `text/template` (embedded in binary) |
| Asset serving | `http.FileServer(http.FS(dist))` via `go:embed` |
| Asset location | `internal/frontend/auth/content/dist` (embedded at compile time) |

## Form Field Names

The PatternFly `LoginForm` component generates specific field IDs that the backend expects:

- Username field: `pf-login-username-id`
- Password field: `pf-login-password-id`

These are parsed in `internal/auth/storage.go:parseCredentials()`.
