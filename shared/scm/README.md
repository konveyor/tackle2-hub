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

`Validate` is also on the interface but is called automatically
by the `New()` factory — callers do not need to invoke it directly.

## Authentication

Both implementations support:

- **Username/password** credentials (Git credential store, SVN password injection).
- **SSH keys** via the `Identity` attached to the `Remote`.
- **Proxy** configuration per scheme (HTTP/HTTPS) with optional proxy authentication.
  Proxies are configured in the hub and fetched automatically by
  the factory — the caller does not need to configure them.

### Username/Password

```go
identity := &api.Identity{
    Name:     "git-creds",
    User:     "admin",
    Password: "changeme",
}
```

### SSH Key

Populate the `Key` field with the private key content. Use
`Password` for the key passphrase when the key is encrypted.

SSH keys are added to the agent using the `shared/ssh` package.
An SSH agent must be running before SCM operations are called.
Importing `shared/ssh` starts an agent automatically via `init()`.
Alternatively, an externally started agent may be used.

```go
import _ "github.com/konveyor/tackle2-hub/shared/ssh"

identity := &api.Identity{
    Name:     "git-ssh",
    Key:      "-----BEGIN OPENSSH PRIVATE KEY-----\n...",
    Password: "passphrase",
}
```

## Factory

Use `New()` to create an SCM instance. The `repository.Kind` field
determines the implementation:

| Kind           | Implementation | Description          |
|----------------|----------------|----------------------|
| `"git"`        | `Git`          | Git repository.      |
| `"svn"`        | `Subversion`   | Subversion (SVN).    |
| `"subversion"` | `Subversion`   | Subversion (SVN).    |

Any unrecognized kind defaults to `Git`.

The factory requires a `binding.RichClient` to fetch hub-managed
configuration from the hub API. Specifically, it uses the client to:

- Read the `git.insecure.enabled` or `svn.insecure.enabled` setting.
- List configured proxies and resolve their associated identities.

These settings are managed centrally in the hub and cannot be derived
from the repository alone.

The `identity` parameter is optional — pass `nil` for public
repositories that do not require authentication.

```go
import (
    "github.com/konveyor/tackle2-hub/shared/api"
    "github.com/konveyor/tackle2-hub/shared/binding"
    "github.com/konveyor/tackle2-hub/shared/binding/auth"
    "github.com/konveyor/tackle2-hub/shared/scm"
)

repository := api.Repository{
    Kind:   "git",
    URL:    "https://github.com/konveyor/tackle2-hub",
    Branch: "main",
}

identity := &api.Identity{
    Name:     "git-creds",
    User:     "admin",
    Password: "changeme",
}

client := binding.New("https://hub.example.com")
bearer := auth.NewBearer("my-token")
client.Client.Use(bearer)

// With credentials.
repo, err := scm.New("/tmp/repo", repository, identity, client)

// Public repository (no credentials).
repo, err := scm.New("/tmp/repo", repository, nil, client)
```

### Repository Fields

| Field    | Description                                                 |
|----------|-------------------------------------------------------------|
| `Kind`   | SCM type (see table above).                                 |
| `URL`    | Remote repository URL.                                      |
| `Branch` | Branch or tag to checkout. Tags are detected automatically. |
| `Tag`    | Not used by this package. Use `Branch` for tags.            |
| `Path`   | SVN-specific sub-path within the repository. Ignored by Git.|

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
repo, _ := scm.New("/tmp/repo", repository, identity, client)

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
