## Settings ##

Settings provide hub configuration.

The `T` specifies the type:
- `I` = integer
- `S` = string
- `B` = boolean (0|1|true|false)

### General ###

| Name                      | T | Envar                 | Default         | Definition                                        |
|---------------------------|---|-----------------------|-----------------|---------------------------------------------------|
| Build                     | S | BUILD                 |                 | Hub build version.                                |
| Namespace                 | S | NAMESPACE             | konveyor-tackle | Home k8s Namespace.                               |
| **DB**.Path               | S | DB_PATH               | /tmp/tackle.db  | Path to sqlite file.                              |
| **DB**.MaxConnections     | I | DB_MAX_CONNECTION     | 1               | Number of DB connections.                         |
| **DB**.SeedPath           | S | DB_SEED_PATH          | /tmp/seed       | Path to seed files.                               |
| **Bucket**.Path           | S | BUCKET_PATH           | /tmp/bucket     | Path to bucket storage directory.                 |
| **Bucket**.TTL            | I | BUCKET_TTL            | 1 (minute)      | Orphaned buckets TTL (minutes).                   |
| **File**.TTL              | I | FILE_TTL              | 1 (minute)      | Orphaned files TTL (minutes).                     |
| **Cache**.RWX             | B | RWX_SUPPORTED         | FALSE           | Cache volume supports RWX.                        |
| **Cache**.Path            | S | CACHE_PATH            | /cache          | Cache volume mount path.                          |
| **Cache**.PVC             | S | CACHE_PVC             | cache           | Cache PVC name. Used when RWX suppored.           |
| **Shared**.Path           | S | SHARED_PATH           | /shared         | Shared volume mount path.                         |
| **Encryption**.Passphrase | S | ENCRYPTION_PASSPHRASE | tackle          | RSA encryption passphrase.                        |
| Development               | B | DEVELOPMENT           | FALSE           | Development mode.                                 |
| Disconnected              | B | DISCONNECTED          | FALSE           | Not connected to a cluster.                       |
| Product                   | S | APP_NAME              | tackle          | Product/application name. Affects target seeding. |
| **Metrics**.Enabled       | B | METRICS_ENABLED       | TRUE            | Metrics reporting enabled.                        |
| **Metrics**.Port          | I | METRICS_PORT          |                 | Metrics reporting (listen) port number.           |

### Auth ###

| Name      | T | Envar   | Default | Definition |
|-----------|---|---------|---------|------------|
| Required  | B |         |         |            |
| RolePath  | S |         |         |            |
| UserPath  | S |         |         |            |
| Token.Key | S |         |         |            |
|           |   |         |         |            |

### Task Manager ###

| Name                    | T | Envar   | Default | Definition |
|-------------------------|---|---------|---------|------------|
| SA                      | S |         |         |            |
| Retries                 | I |         |         |            |
| Reaper.Created          | I |         |         |            |
| Reaper.Succeeded        | I |         |         |            |
| Reaper.Failed           | I |         |         |            |
| Preemption.Enabled      | B |         |         |            |
| Preemption.Delayed      | I |         |         |            |
| Preemption.Postponed    | I |         |         |            |
| Preemption.Rate         | I |         |         |            |
| Pod.Retention.Succeeded | I |         |         |            |
| Pod.Retention.Failed    | I |         |         |            |
| Pod.UID                 | S |         |         |            |

### Intervals/Frequencies ###

| Name   | T | Envar   | Default | Definition |
|--------|---|---------|---------|------------|
| Task   | I |         |         |            |
| Reaper | I |         |         |            |

### Analysis ###

| Name            | T | Envar   | Default | Definition |
|-----------------|---|---------|---------|------------|
| ReportPath      | S |         |         |            |
| ArchiverEnabled | B |         |         |            |

### Discovery ###

| Name    | T | Envar   | Default | Definition                                    |
|---------|---|---------|---------|-----------------------------------------------|
| Enabled | B |         |         | Trigger discover tasks on application update. |
| Label   | S |         |         | k8s label use to select lang-discover tasks.  |

### Addon ###

| Name      | T | Envar   | Default | Definition |
|-----------|---|---------|---------|------------|
| HomeDir   | S |         |         |            |
| Hub.URL   | S |         |         |            |
| Hub.Token | S |         |         |            |
| Task.ID   | I |         |         |            |


### KEYCLOAK ###

| Name                   | T | Envar   | Default | Definition |
|------------------------|---|---------|---------|------------|
| RequirePasswordUpdate  | B |         |         |            |
| Host                   | S |         |         |            |
| Realm                  | S |         |         |            |
| ClientID               | S |         |         |            |
| ClientSecret           | S |         |         |            |
| Admin.User             | S |         |         |            |
| Admin.Password         | S |         |         |            |
| Admin.Realm            | S |         |         |            |
