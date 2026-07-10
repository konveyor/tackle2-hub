# Frontend Components

This directory contains frontend UI components that are embedded into the hub binary.

## Organization

Each frontend app lives in its own subdirectory:

```
internal/frontend/
├── auth/           # OIDC login and device authorization pages
│   ├── handler.go  # Go handler serving embedded assets
│   └── content/    # React + PatternFly frontend project
└── pkg.go          # Handler registration
```

## General Architecture

Frontend apps in this directory:

- **Are embedded** into the Go binary at compile time via `go:embed`
- **Are served** by Go handlers that implement the `Handler` interface
- **Are built** using modern JavaScript tooling (React, TypeScript, Rspack)
- **Are standalone** - each app has its own build configuration and dependencies

## Adding a New Frontend App

1. **Create directory structure:**
   ```
   internal/frontend/myapp/
     handler.go
     content/
       package.json
       src/
       dist/
   ```

2. **Implement handler in `handler.go`:**
   ```go
   package myapp
   
   import (
       "embed"
       "io/fs"
       "net/http"
       "github.com/gin-gonic/gin"
   )
   
   const Route = "/frontend/myapp"
   
   //go:embed content/dist
   var content embed.FS
   
   var dist fs.FS
   
   func init() {
       var err error
       dist, err = fs.Sub(content, "content/dist")
       if err != nil {
           panic(err)
       }
   }
   
   type Handler struct {}
   
   func (h Handler) AddRoutes(e *gin.Engine) {
       assetHandler := http.StripPrefix(Route, http.FileServer(http.FS(dist)))
       routeGroup := e.Group(Route)
       routeGroup.GET("/*path", func(c *gin.Context) {
           assetHandler.ServeHTTP(c.Writer, c.Request)
       })
   }
   ```

3. **Register in `pkg.go`:**
   ```go
   func ALL() []Handler {
       return []Handler{
           auth.Handler{},
           myapp.Handler{},  // Add here
       }
   }
   ```

4. **Build tooling will automatically discover it** if using the convention-based Makefile pattern.

## Building

Frontend apps are built by the `make frontend` target, which discovers and builds all apps in `internal/frontend/*/content/`:

```bash
make frontend       # Build all frontend apps
make clean-frontend # Clean all frontend builds
```

## See Also

- `internal/frontend/auth/README.md` - Login page frontend documentation
- `Makefile` - Build targets for frontend apps
