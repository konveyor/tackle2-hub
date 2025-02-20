## Settings ##

Settings provide hub configuration.

The `T` specifies the type:
- `I` = integer
- `S` = string
- `B` = boolean (0|1|true|false)

### General ###

| Name                      | T | Envar   | Default | Definition |
|---------------------------|---|---------|---------|------------|
| Build                     | S |         |         |            |
| Namespace                 | S |         |         |            |
| **DB**.Path               | S |         |         |            |
| **DB**.MaxConnections     | I |         |         |            |
| **DB**.SeedPath           | S |         |         |            |
| **Bucket**.Path           | S |         |         |            |
| **Bucket**.TTL            | I |         |         |            |
| **File**.TTL              | I |         |         |            |
| **Cache**.RWX             | B |         |         |            |
| **Cache**.Path            | S |         |         |            |
| **Cache**.PVC             | S |         |         |            |
| **Shared**.Path           | S |         |         |            |
| **Encryption**.Passphrase | S |         |         |            |
| Development               | B |         |         |            |
| Disconnected              | B |         |         |            |
| Product                   | S |         |         |            |
| **Metrics**.Enabled       | B |         |         |            |
| **Metrics**.Port          | I |         |         |            |

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
