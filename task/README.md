
## Manager ##

### Flow ###

### Priority ###

### Preemption ###

### Macros ###

## Task ##

### Properties ###

`*` indicates reported by addon.

| Name        | Definition                                                                                                         |
|-------------|--------------------------------------------------------------------------------------------------------------------|
| ID          | Unique identifier.                                                                                                 |
| CreateTime  | The timestamp of when the task was created.                                                                        |
| CreateUser  | The user (name) that created the task.                                                                             |
| UpdateUser  | The user (name) that last updated the task.                                                                        |
| Name        | The task mame (non-unique).                                                                                        |
| Kind        | The kind references a Task (kind) CR by name.                                                                      |
| Addon       | The addon to be executed. References an Addon CR by name. When not specified, the addon is selected based on kind. |
| Extension   | The list of extensions to be injected into the addon pod as _sidecar_ containers.                                  |
| State       | The task state.  See: _States_.                                                                                    |
| Locator     | The task locator. An arbitrary user-defined value used for lookup.                                                 |
| Priority    | The task execution priority. See: _Priority_.                                                                      |
| Policy      | The task execution policy. Determines when task is postponed. See: _Policy_.                                       |
| TTL         | The task Time-To-Live in each state. See: _TTL_.                                                                   |
| Data        | The data provided to the addon. The schema is dictated by each addon. This may be _ANY_ document.                  |
| Started     | The UTC timestamp when the task execution started.                                                                 |
| Terminated  | The UTC timestamp when execution completed.                                                                        |
| Errors      | A list of reported errors. See: _Errors_.                                                                          |
| Events      | A list of reported task processing events. See: _Events_.                                                          |
| Pod         | The fully qualified name of the pod created.                                                                       |
| Retries     | The number of times failure to create a pod is retried. This does not include when blocked by resource quota.      |
| Attached    | Files attached to the task.                                                                                        |
| \*Activity  | The activity (log) entries are reported by the addon. Intended to reflect what the addon is doing.                 |
| \*Total     | Progress: The total number of items to be completed by the addon.                                                  |
| \*Completed | Progress: The number of items completed by the addon.                                                              | |

### Events ###

### Errors ###

### States ###

| State        | Definition                                                                                    |
|:-------------|:----------------------------------------------------------------------------------------------|
| Created      | The task has been created but not submitted.                                                  |
| Ready        | The task has been submitted to the manager and will be scheduled for execution.               |
| Postponed    | The task has been postponed until another task has completed based on task scheduling _policy_. |
| QuotaBlocked | The task pod has been (temporarily) prevented from being created by k8s resource quota.       |
| Pending      | The task pod has been created and awaiting k8s scheduling.                                    |
| Running      | The task pod is running.                                                                      |
| Succeeded    | The task pod successfully completed.                                                          |
| Failed       | The task pod either failed to be started by k8s or completed with errors.                     |
| Canceled     | The task has been canceled.                                                                   |


### Policy ###

### TTL (Time-To-Live) ###

#### Isolation ####

#### PreemptEnabled ####

#### PreemptExempt ####

## (Task) Kinds ###

## Addons ##

### Selection ###

## Extensions ##

### Selection ###

## Authorization ##


