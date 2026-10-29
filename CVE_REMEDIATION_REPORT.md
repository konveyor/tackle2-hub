# CVE Remediation Report

## Scope

| Ecosystem | Manifest | Scanned with |
|-----------|----------|--------------|
| Go (main module) | `go.mod` | govulncheck, trivy, grype |
| Go (`shared` module) | `shared/go.mod` | govulncheck, trivy, grype |
| npm (login page) | `internal/frontend/auth/content/package-lock.json` | trivy, grype |

Tools: `govulncheck ./...` (run at the repo root and in `shared/`), `trivy fs --scanners vuln .`, `grype dir:.`.

## Findings fixed

| Advisory / CVE | Package | Severity (highest reported) | From | To | Module(s) |
|---|---|---|---|---|---|
| GO-2026-6603, GO-2026-6610, GO-2026-6611, GO-2026-6612, GO-2026-6617 (CVE-2026-78659, CVE-2026-78660, CVE-2026-78663, CVE-2026-78669, CVE-2026-97032) | golang.org/x/net | Critical | v0.58.0 / v0.57.0 | v0.62.0 | root, shared |
| GO-2026-6860, GO-2026-6866 (CVE-2026-78661, CVE-2026-97045) | golang.org/x/net | Unknown | v0.58.0 / v0.57.0 | v0.62.0 | root, shared |
| GO-2026-6354, GO-2026-6355 (CVE-2026-56855, CVE-2026-78662) | golang.org/x/crypto | High | v0.55.0 | v0.58.0 (required by x/net v0.62.0) | root |
| GO-2026-6629 (CVE-2026-56851) | golang.org/x/text | High | v0.40.0 | v0.43.0 (required by x/net v0.62.0) | shared |
| GHSA-mgcc-cww5-fp2q (CVE-2026-116587) | github.com/go-jose/go-jose/v4 | High | v4.1.4 | v4.1.6 | root |
| GHSA-vvm8-gf78-28g4 (CVE-2026-116585), GHSA-4w3g-jfj2-cw2q (CVE-2026-116586) | github.com/go-jose/go-jose/v4 | Medium | v4.1.4 | v4.1.6 | root |
| GHSA-mh63-6h87-95cp / GO-2025-3553 (CVE-2025-30204) | github.com/golang-jwt/jwt/v4 (indirect, via go-jira) | High | v4.5.0 | v4.5.2 | root |
| GHSA-29wx-vh33-7x7r / GO-2024-3250 (CVE-2024-51744) | github.com/golang-jwt/jwt/v4 | Low | v4.5.0 | v4.5.2 | root |
| GHSA-pjcq-xvwq-hhpj / GO-2026-5543 (CVE-2026-32952) | github.com/Azure/go-ntlmssp (indirect, via go-ldap) | Medium | v0.1.0 | v0.1.1 | root |
| GO-2026-5774, GO-2026-5775, GO-2026-5777 | github.com/go-chi/chi/v5 (indirect, via zitadel/oidc) | Unknown | v5.2.5 | v5.4.0 | root |
| GHSA-j6qf-3m9c-vwj3 (CVE-2026-114437) | github.com/go-chi/chi/v5 | Medium | v5.2.5 | v5.4.0 | root |
| GHSA-259r-337f-4rfw | github.com/klauspost/compress (indirect, via prometheus) | High | v1.18.0 | v1.18.7 | root |
| GO-2026-6179, GO-2026-6180 | golang.org/x/mod (indirect, via swag / x/tools) | High | v0.38.0 | v0.41.0 (pulled in by x/tools v0.51.0 when x/net was bumped) | root |

govulncheck had flagged five of the x/net advisories as reachable from project code, in `internal/api/service.go` (the reverse proxy), `internal/k8s` and `internal/tracker`. All of them are now fixed.

The npm lockfile had no findings in either trivy or grype.

### Side effects to review

- Moving to `golang.org/x/net` v0.62.0 also raised `golang.org/x/sys` (v0.49.0), `golang.org/x/sync` (v0.24.0), `golang.org/x/term` (v0.47.0) and `golang.org/x/tools` (v0.51.0) through minimum version selection.
- x/net v0.62.0 requires Go 1.26, so the `go` line in `shared/go.mod` went from `1.25.0` to `1.26.0`. The root module already declares `go 1.26.2`. Anyone who uses the `shared` module from outside this repo now needs Go 1.26 or newer.
- I bumped `golang.org/x/mod` on its own in commit `980a17f`, and then reverted it in `cabe609`. `go get golang.org/x/mod@v0.40.0` downgraded x/net, x/crypto, x/text and x/tools, which would have brought back the vulnerabilities fixed earlier. It also rewrote the `shared` require version. The revert is correct: x/mod is already at v0.41.0, which clears both advisories.

## Unresolved

| Advisory | Package | Reason not fixed | Suggested manual action |
|---|---|---|---|
| GO-2026-5932 | golang.org/x/crypto (v0.58.0), the `openpgp` package | There is no fixed version. The advisory says `golang.org/x/crypto/openpgp` is unmaintained and unsafe by design, so it applies to every version of the module. The project doesn't use the package: `go mod why golang.org/x/crypto/openpgp` says the main module does not need it, `go list -deps ./...` doesn't include it, and govulncheck reports it as not called. | Nothing to do now. Don't import `golang.org/x/crypto/openpgp`; use `github.com/ProtonMail/go-crypto/openpgp` if OpenPGP is ever needed. Scanners that work at module level (trivy, grype) will keep listing it until the advisory is withdrawn or scoped down. |

## Verification

- `go build ./...` and `go vet ./internal/... ./cmd/...` pass in the root module after every bump. `go build ./...` passes in `shared/`.
  - Building needs `internal/frontend/auth/content/dist` (an `//go:embed` target that is gitignored and produced by the npm build). For local checks I used a temporary placeholder `index.html` and deleted it afterwards.
- Final re-scan:
  - `govulncheck ./...` (root and `shared/`): **No vulnerabilities found.**
  - `trivy fs --scanners vuln .`: only GO-2026-5932.
  - `grype dir:.`: only GO-2026-5932.
- `go test` shows 6 failing tests, and they already fail on the starting commit `680736a` (I checked in a separate worktree), so these upgrades did not cause them:
  - `internal/auth`: `TestSeedClientsFromCRD`, `TestSeedClientsUpdate`, `TestSeedClientsDeleteOrphaned`, `TestSeedClientsIDPreservation`
  - `shared/binding/auth`: `TestOIDCHeader`, `TestOIDCTokenMethods`
