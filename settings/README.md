## Settings ##

Settings provide hub configuration.

The `T` specifies the type:
- `I` = integer
- `S` = string
- `B` = boolean (0|1|true|false)

### Main ###

These settings pertain to the Hub in general.

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

### Logging ###

Logging verbosity (0=Lowest).
For all components:
- 0 will log SQL slower than 1 second.
- 1+ all SQL is logged.

| Name          | T | Envar         | Default | Description                         |
|---------------|---|---------------|---------|-------------------------------------|
| Log.Master    | I | LOG_MASTER    | 0       | Master logging verbosity.           |
| Log.Migration | I | LOG_MIGRATION | 0       | Migrations logging verbosity        |
| Log.Web       | I | LOG_WEB       | 0       | Web/REST API logging verbosity.     |
| Log.Reaper    | I | LOG_REAPER    | 0       | Reaper logging verbosity            |
| Log.Task      | I | LOG_TASK      | 0       | Task manager logging verbosity.     |
| Log.Auth      | I | LOG_AUTH      | 0       | Auth logging verbosity.             |
| Log.Heap      | I | LOG_HEAP      | 0       | Heap information logging verbosity. |
| Log.Command   | I | LOG_CMD       | 0       | Command (shell) logging verbosity.  |
| Log.SSH       | I | LOG_SSH       | 0       | SSH agent logging verbosity.        |
| Log.SCM       | I | LOG_SCM       | 0       | SCM logging verbosity.              |

### Auth ###

These settings pertain to authentication and authorization.

| Name      | T | Envar         | Default         | Definition                                 |
|-----------|---|---------------|-----------------|--------------------------------------------|
| Required  | B | AUTH_REQUIRED | FALSE           | API enforces authentication/authorization. |
| RolePath  | S | RULE_PATH     | /tmp/roles.yaml | Path to file used to seed roles.           |
| UserPath  | S | USER_PATH     | /tmp/users/yaml | Path to file used to seed users.           |
| Token.Key | S | ADDON_TOKEN   |                 | Key used to sign tokens.                   |

### Task Manager ###

These settings pertain to the tasking system.

| Name                    | T | Envar                     | Default    | Definition                                                              |
|-------------------------|---|---------------------------|------------|-------------------------------------------------------------------------|
| Enabled                 | B | TASK_ENABLED              | TRUE       | Tasking enabled. FALSE when Disconnected=TRUE.                          |
| SA                      | S | TASK_SA                   |            | Task pod service account name.                                          |
| Retries                 | I | TASK_RETRIES              | 1          | Task pod creation retires.                                              |
| Reaper.Created          | I | TASK_REAP_CREATED         | 72 (hour)  | (seconds) task may remain in state=CREATED before deleted.              |
| Reaper.Succeeded        | I | TASK_REAP_SUCCEEDED       | 72 (hour)  | (seconds) before SUCCEEDED task's bucket released.                      |
| Reaper.Failed           | I | TASK_REAP_FAILED          | 720 (hour) | (seconds) before FAILED task's bucket released.                         |
| Pod.Retention.Succeeded | I | TASK_POD_RETAIN_SUCCEEDED | 1 (minute) | (minutes) before SUCCEEDED task pod is reaped (deleted).                |
| Pod.Retention.Failed    | I | TASK_POD_RETAIN_FAILED    | 72 (hour)  | (minutes) before FAILED task pod is reaped (deleted).                   |
| Pod.UID                 | S | TASK_UID                  |            | Task pod run-as user id.                                                |
| Pod.Quota               | I | TASK_POD_QUOTA            | 20         | Task pod quota. (0=unlimited). May be overridden by k8s resource quota. |


### Intervals/Frequencies ###

These settings pertain to the frequency of _manager_ main loops.

| Name   | T | Envar            | Default    | Definition                           |
|--------|---|------------------|------------|--------------------------------------|
| Task   | I | FREQUENCY_TASK   | 1 (second) | (seconds) between each manager pass. |
| Reaper | I | FREQUENCY_REAPER | 1 (minute) | (minutes) between each reaper pass.  |

### Analysis ###

These settings pertain to analysis.

| Name            | T | Envar                     | Default               | Definition                      |
|-----------------|---|---------------------------|-----------------------|---------------------------------|
| ReportPath      | S | ANALYSIS_REPORT_PATH      | /tmp/analysis/report  | Path to static analysis report. |
| ArchiverEnabled | B | ANALYSIS_ARCHIVER_ENABLED | TRUE                  | Analysis report auto-archived.  |

### Discovery ###

These settings pertain to the auto-create of lang-discovery tasks.

| Name    | T | Envar             | Default                    | Definition                                    |
|---------|---|-------------------|----------------------------|-----------------------------------------------|
| Label   | S | DISCOVERY_LABEL   | konveyor.io/discovery      | k8s label use to select lang-discover tasks.  |

### Addon ###

These settings are intended to be shared by the hub and the (Go) addons.

| Name      | T | Envar        | Default               | Definition                      |
|-----------|---|--------------|-----------------------|---------------------------------|
| HomeDir   | S | ADDON_HOME   | /addon                | Addon home (working) directory. |
| Hub.URL   | S | HUB_BASE_URL | http://localhost:8080 | Hub (base) URL.                 |
| Hub.Token | S | TOKEN        |                       | Auth token for hub API.         |
| Task.ID   | I | TASK         |                       | Task addon working on.          |


### KEYCLOAK ###

| Name                   | T | Envar                    | Default | Definition                 |
|------------------------|---|--------------------------|---------|----------------------------|
| RequirePasswordUpdate  | B | KEYCLOAK_REQ_PASS_UPDATE | TRUE    | User must change password. |
| Host                   | S | KEYCLOAK_HOST            |         | Service hostname.          |
| Realm                  | S | KEYCLOAK_REALM           |         | Realm name.                |
| ClientID               | S | KEYCLOAK_CLIENT_ID       |         | Client id.                 |
| ClientSecret           | S | KEYCLOAK_CLIENT_SECRET   |         | Client secret.             |
| Admin.User             | S | KEYCLOAK_ADMIN_USER      |         | Admin client user.         |
| Admin.Password         | S | KEYCLOAK_ADMIN_PASS      |         | Admin client password.     |
| Admin.Realm            | S | KEYCLOAK_ADMIN_REALM     |         | Admin client realm.        |
