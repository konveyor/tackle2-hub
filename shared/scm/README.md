# SCM Package

The `scm` package provides a unified interface for working with
SCM (Source Configuration Management) repositories. It supports
Git and Subversion through a common `SCM` interface.

## Prerequisites

This package shells out to the `git` and `svn` command-line tools.
Both must be installed and available on `PATH`.

## Interface

The `SCM` interface defines repository operations:

| Method     | Description                                    |
|------------|------------------------------------------------|
| `Fetch`    | Clone the repository.                          |
| `Update`   | Pull latest changes from the remote.           |
| `Branch`   | Switch to a branch, optionally creating it.    |
| `Commit`   | Stage files, commit, and push.                 |
| `Push`     | Push committed changes to the remote.          |
| `Head`     | Return the current HEAD commit/revision.       |
| `Clean`    | Delete working files created by the SCM.       |

`Validate` checks the remote configuration (for example, rejecting an
`http` URL unless `Insecure` is set). It is **not** called by `New()`;
invoke it explicitly after constructing the instance.

`WithProxies` configures HTTP/HTTPS proxies after construction (see
[Proxies](#proxies)).

## Authentication

Both implementations support:

- **Username/password** credentials (Git credential store, SVN password injection).
- **SSH keys** via the `Identity` attached to the `Remote`.
- **Proxy** configuration per scheme (HTTP/HTTPS) with optional proxy authentication,
  applied via `WithProxies` (see [Proxies](#proxies)).

Credentials are supplied through the `Identity` field on the `Remote`.
`Identity` is optional — leave it `nil` for public repositories that
do not require authentication.

### Username/Password

```go
identity := &scm.Identity{
    Name:     "git-creds",
    User:     "admin",
    Password: "changeme",
}
```

### SSH Key

Populate the `Key` field with the private key content. Use
`Password` for the key passphrase when the key is encrypted.

An SSH agent is started automatically — both the Git and Subversion
implementations import `shared/ssh`, which starts an agent via
`init()`.

```go
identity := &scm.Identity{
    Name:     "git-ssh",
    Key:      "-----BEGIN OPENSSH PRIVATE KEY-----\n...",
    Password: "passphrase",
}
```

## Factory

Use `New()` to create an SCM instance from a `Remote`. The
`Remote.Kind` field determines the implementation:

| Kind           | Implementation | Description          |
|----------------|----------------|----------------------|
| `"git"`        | `Git`          | Git repository.      |
| `"svn"`        | `Subversion`   | Subversion (SVN).    |
| `"subversion"` | `Subversion`   | Subversion (SVN).    |

Any unrecognized kind defaults to `Git`.

`New()` is self-contained: it depends only on package-native types
(`Remote`, `Identity`) and does **not** contact the hub. This lets
external tools use the package without hub connectivity or settings.
`New()` only constructs the instance — call `Validate()` afterward,
and `WithProxies()` if proxies are needed.

```go
import (
    "github.com/konveyor/tackle2-hub/shared/scm"
)

remote := scm.Remote{
    Kind:   "git",
    URL:    "https://github.com/konveyor/tackle2-hub",
    Branch: "main",
    Identity: &scm.Identity{
        Name:     "git-creds",
        User:     "admin",
        Password: "changeme",
    },
}

repo := scm.New("/tmp/repo", remote)
err := repo.Validate()

// Public repository (no credentials): leave Identity nil.
remote.Identity = nil
repo = scm.New("/tmp/repo", remote)
```

### Remote Fields

| Field      | Description                                                   |
|------------|---------------------------------------------------------------|
| `Kind`     | SCM type (see table above).                                   |
| `URL`      | Remote repository URL.                                        |
| `Branch`   | Branch or tag to checkout. Tags are detected automatically.   |
| `Path`     | SVN-specific sub-path within the repository. Ignored by Git.  |
| `Identity` | Optional credentials (see [Authentication](#authentication)). |
| `Insecure` | Allow an `http` URL / skip server certificate verification.   |

## Proxies

`WithProxies` configures the HTTP/HTTPS proxies used for remote
operations. It takes a `ProxyMap` keyed by scheme (`"http"` /
`"https"`) and must be called after `New()` and before the operation
that needs it. Proxies are optional — skip `WithProxies` entirely
when none are required.

```go
proxies := scm.ProxyMap{
    "https": {
        Kind:     "https",
        Host:     "proxy.example.com",
        Port:     8080,
        Excluded: []string{"internal.example.com"},
        Identity: &scm.Identity{
            User:     "proxyuser",
            Password: "secret",
        },
    },
}

repo := scm.New("/tmp/repo", remote)
repo.WithProxies(proxies)
_ = repo.Validate()
```

### Proxy Fields

| Field      | Description                                                   |
|------------|---------------------------------------------------------------|
| `Kind`     | Scheme the proxy serves (`"http"` or `"https"`).              |
| `Host`     | Proxy host.                                                   |
| `Port`     | Proxy port.                                                   |
| `Excluded` | Hosts that bypass the proxy.                                  |
| `Identity` | Optional proxy credentials (see [Authentication](#authentication)). |

An external tool can build a `ProxyMap` itself, as above. When
running against the hub, the `Proxies` helper builds the map from the
hub's proxy list (see [Hub-managed configuration](#hub-managed-configuration)).

## Hub-managed configuration

The insecure setting and proxy list are managed centrally in the hub
and cannot be derived from the remote alone. The package provides
optional helpers that read them via a `binding.RichClient`. Both take
hub types (`api.Repository`, `binding.RichClient`) rather than the
package-native `Remote`:

- `Insecure(client, repository)` — returns the `git.insecure.enabled`
  or `svn.insecure.enabled` setting for the `repository.Kind`.
- `Proxies(client)` — returns the enabled proxies keyed by scheme,
  with their identities resolved.

These are opt-in: an external tool that does not use the hub can skip
them entirely and set `Remote.Insecure` (and, if needed, proxies via
`WithProxies`) itself. Addons run inside the hub and typically use the
`shared/addon/scm` wrapper, which maps the `api.Repository`/`api.Identity`
to a `Remote` and applies both helpers automatically.

```go
// repository is the hub's api.Repository resource.
insecure, err := scm.Insecure(client, repository)

remote := scm.Remote{
    Kind:     repository.Kind,
    URL:      repository.URL,
    Branch:   repository.Branch,
    Path:     repository.Path,
    Insecure: insecure,
}

repo := scm.New("/tmp/repo", remote)

proxies, err := scm.Proxies(client)
repo.WithProxies(proxies)

err = repo.Validate()
```

## Typical Workflow

`Fetch` must be called first to clone the repository. The `destDir`
must either not exist or be an empty directory — `Fetch` will fail
on a non-empty directory. It should only be called once; use `Update`
to pull new changes afterward.

After `Fetch`, methods like `Branch`, `Commit`, `Push`, `Update`,
and `Head` operate on the cloned working directory. `Branch` can be
called multiple times to switch between branches on the same instance.

Call `Clean` when finished to remove credentials and configuration.
The cloned repository at `destDir` is not removed — the caller is
responsible for cleaning that up.

```go
repo := scm.New("/tmp/repo", remote)
_ = repo.Validate()

err = repo.Fetch()

err = repo.Branch("my-feature", scm.CREATE)

// ... write files into /tmp/repo ...

err = repo.Commit([]string{"output.yaml"}, "add results")

commit, err := repo.Head()

err = repo.Clean()
```

## Methods

### Fetch

Clones the repository into the destination directory. The `destDir`
must not exist or be empty. When a branch is configured on the
remote, it is checked out after cloning. `Fetch` should only be
called once per instance — use `Update` to refresh.

```go
err := repo.Fetch()
```

Equivalent git:

```
git clone <url> /tmp/repo
git checkout -B <branch> origin/<branch>
```

### Update

Fetches the latest refs and pulls commits for the current branch.
Tags are fetched and stale remote-tracking refs are pruned.

```go
err := repo.Update()
```

Equivalent git:

```
git fetch --tags --prune
git pull
```

### Branch

Switches to the named branch. The branch is resolved in order:
tag, remote-tracking branch (`origin/<ref>`), then local branch.

```go
err := repo.Branch("v1.0")
```

Equivalent git:

```
git fetch --tags --prune
git checkout -B v1.0 origin/v1.0
```

When `CREATE` is passed and checkout fails, a new branch is created
from the current HEAD and pushed upstream. `CREATE` is a bitmask
option — multiple options can be passed as variadic arguments.

```go
err := repo.Branch("my-feature", scm.CREATE)
```

Equivalent git:

```
git checkout -b my-feature
git push -u origin my-feature
```

### Commit

Stages the specified files, creates a commit, and pushes to the
remote. The `--allow-empty` flag is used so commits succeed even
when no files have changed. Because `Commit` pushes automatically,
a separate `Push` call is not needed after `Commit`.

If the commit succeeds but the push fails, the local commit exists
but is not on the remote. The error is returned, and a subsequent
`Push()` call can recover.

```go
err := repo.Commit(
    []string{"report.yaml", "output/data.json"},
    "updated analysis")
```

Equivalent git:

```
git add report.yaml output/data.json
git commit --allow-empty -m "updated analysis"
git push origin HEAD
```

### Push

Pushes the current HEAD to the remote. This is useful when commits
are made outside of the `Commit` method (e.g., by an external tool
writing into the working directory).

```go
err := repo.Push()
```

Equivalent git:

```
git push origin HEAD
```

### Head

Returns the SHA of the current HEAD commit.

```go
commit, err := repo.Head()
```

Equivalent git:

```
git rev-parse HEAD
```

### Clean

Deletes the SCM home directory (credentials, config) created during
operations. The cloned repository at `destDir` is not removed — the
caller is responsible for cleaning that up.

```go
err := repo.Clean()
```
