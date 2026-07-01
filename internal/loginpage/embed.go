// Package loginpage embeds the compiled login page frontend assets.
// The assets are built from the login-page/ project at the repository root
// by running `npm run build` in that directory. Build output is written
// directly to internal/loginpage/dist/ so that this go:embed picks them up.
package loginpage

import "embed"

// FS contains the compiled login page dist/ output.
//
//go:embed dist
var FS embed.FS
