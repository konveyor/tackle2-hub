## Settings ##

### General ###

| Name                  | Envar   | Default | Definition |
|-----------------------|---------|---------|------------|
| Build                 |         |         |            |
| Namespace             |         |         |            |
| DB.Path               |         |         |            |
| DB.MaxConnections     |         |         |            |
| DB.SeedPath           |         |         |            |
| Bucket.Path           |         |         |            |
| Bucket.TTL            |         |         |            |
| File.TTL              |         |         |            |
| Cache.RWX             |         |         |            |
| Cache.Path            |         |         |            |
| Cache.PVC             |         |         |            |
| Shared.Path           |         |         |            |
| Encryption.Passphrase |         |         |            |
| Development           |         |         |            |
| Disconnected          |         |         |            |
| Product               |         |         |            |
| Metrics.Enabled       |         |         |            |
| Metrics.Port          |         |         |            |
|                       |         |         |            |

### Auth ###

| Name      | Envar   | Default | Definition |
|-----------|---------|---------|------------|
| Required  |         |         |            |
| RolePath  |         |         |            |
| UserPath  |         |         |            |
| Token.Key |         |         |            |
|           |         |         |            |

### Task Manager ###

| Name                    | Envar   | Default | Definition |
|-------------------------|---------|---------|------------|
| SA                      |         |         |            |
| Retries                 |         |         |            |
| Reaper.Created          |         |         |            |
| Reaper.Succeeded        |         |         |            |
| Reaper.Failed           |         |         |            |
| Preemption.Enabled      |         |         |            |
| Preemption.Delayed      |         |         |            |
| Preemption.Postponed    |         |         |            |
| Preemption.Rate         |         |         |            |
| Pod.Retention.Succeeded |         |         |            |
| Pod.Retention.Failed    |         |         |            |
| Pod.UID                 |         |         |            |

### Intervals/Frequencies ###

| Name   | Envar   | Default | Definition |
|--------|---------|---------|------------|
| Task   |         |         |            |
| Reaper |         |         |            |

### Analysis ###

| Name            | Envar   | Default | Definition |
|-----------------|---------|---------|------------|
| ReportPath      |         |         |            |
| ArchiverEnabled |         |         |            |

### Discovery ###

| Name    | Envar   | Default | Definition                                    |
|---------|---------|---------|-----------------------------------------------|
| Enabled |         |         | Trigger discover tasks on application update. |
| Label   |         |         | k8s label use to select lang-discover tasks.  |

### Addon ###

| Name      | Envar   | Default | Definition |
|-----------|---------|---------|------------|
| HomeDir   |         |         |            |
| Hub.URL   |         |         |            |
| Hub.Token |         |         |            |
| Task.ID   |         |         |            |


### KEYCLOAK ###

| Name                   | Envar   | Default | Definition |
|------------------------|---------|---------|------------|
| RequirePasswordUpdate  |         |         |            |
| Host                   |         |         |            |
| Realm                  |         |         |            |
| ClientID               |         |         |            |
| ClientSecret           |         |         |            |
| Admin.User             |         |         |            |
| Admin.Password         |         |         |            |
| Admin.Realm            |         |         |            |
