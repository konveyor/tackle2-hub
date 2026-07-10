# tackle2-hub

[![Hub Repository on Quay](https://quay.io/repository/konveyor/tackle2-hub/status "Hub Repository on Quay")](https://quay.io/repository/konveyor/tackle2-hub) [![License](http://img.shields.io/:license-apache-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0.html) [![contributions welcome](https://img.shields.io/badge/contributions-welcome-brightgreen.svg?style=flat)](https://github.com/konveyor/tackle2-hub/pulls) [![Hub main CI](https://github.com/konveyor/tackle2-hub/actions/workflows/main.yml/badge.svg?branch=main)](https://github.com/konveyor/tackle2-hub/actions/workflows/main.yml) [![Hub Test Suite nightly](https://github.com/konveyor/tackle2-hub/actions/workflows/test-nightly.yml/badge.svg?branch=main)](https://github.com/konveyor/tackle2-hub/actions/workflows/test-nightly.yml)

Tackle (2nd generation) hub component.

<img src="https://github.com/konveyor/tackle2-hub/blob/main/arch.png" width="850" height="600">

See [settings](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#settings)
for configuration.

The hub provides a REST API, inventory and
[task manager](https://github.com/konveyor/tackle2-hub/blob/main/task/README.md#manager).

## Development

### Prerequisites

- Go (see `go.mod` for the required version)
- Node.js >= 22 and npm >= 10 (required only when working on the login page)

### Building

Build the hub binary **and** the login page frontend in one step:

```bash
make hub
```

### Running locally with the login page

The hub reads login page assets from `LOGIN_PAGE_PATH` (default `/opt/app/login-page`).
The `run` target builds the frontend, then runs the hub binary with `LOGIN_PAGE_PATH`
automatically pointed at the local build output (`login-page/dist`):

```bash
make run
```

To override the path (e.g. a pre-built copy elsewhere):

```bash
make run LOGIN_PAGE_PATH=/path/to/login-page/dist
```

### Building the login page separately

```bash
make login-page
```

This runs `npm ci` (only when `node_modules/` is absent or `package-lock.json` changes)
followed by `npm run build`. Output goes to `login-page/dist/`.

To clean login page build artifacts:

```bash
make clean-login-page
```

### Container builds

The container build handles the login page through a dedicated builder stage in the
[Dockerfile](Dockerfile) and does not require a local Node.js installation:

```bash
make podman-build   # or make docker-build
```

## Code of Conduct
Refer to Konveyor's Code of Conduct [here](https://github.com/konveyor/community/blob/main/CODE_OF_CONDUCT.md).
