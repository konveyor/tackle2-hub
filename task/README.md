
## Manager ##

### Processing ###

The manager processes tasks at an interval defined by the
[Frequency.Task](https://github.com/konveyor/tackle2-hub/tree/main/settings#intervalsfrequencies) setting. 
1. Fetch cluster resources using a k8s cached client.
2. Process queued task delete and cancel requests.
3. Delete orphaned pods. Orphans are pod within the namespace with task 
   labels that does not correspond to a task in the running state.
4. Fetch running tasks; update their status based on the associated pod/container status.
5. Adjust estimated cluster capacity.
6. Kill zombies. Zombies are _sidecar_ containers that have not terminated on their own after the
   _main_ (addon) container has terminated.
7. Fetch and run new (state=Ready) tasks:
   1. select addon. See: _Addons.Selection_.
   2. select extensions. See: _Extensions.Selection_.
   3. create pod. See: _Pods_.

### Priority ###

Tasks with state=Ready are started based on their `Priority` property.
Priority zero(0) is the lowest and the default. There is no maximum. The manager process tasks ordered by
priority. As a result, task pods are created in the order of priority. However, after the pod is created,
the pod scheduling order is at the discretion of the k8s node-scheduler. To maximize the influence of task
priority ordering, it is highly recommended for administrators to create a k8s _Resource Quota_ in the
namespace to restrict the number of pods created.
Priority: 0-9 are reserved for system tasks. User defined tasks created with a priority < 10 are
adjusted to be the priority+10.  This preserves the intended priority ordering within the user range.
For example: A task created without the priority specified (default:0) will be adjusted to have
a priority=10.  A task created with priority=4 will be adjusted to priority=14.

### Quotas ###

The task manager supports an _internal_ quota (setting) to restrict the number of task pods
Pending or  Running at any given time. This is more effective than k8s ResourceQuotas because it
differentiates between scheduled and completed pods. The pod quota _setting_ is ignored when a
k8s ResourceQuota exists in the namespace.
When a pod cannot be created due to quota restriction, the manager sets the state=_QuotaBlocked_ 
and a _QuotaBlocked_ event is reported on the task. The manager will attempt to create the pod in
every processing cycle.

### Cluster capacity detection ###
Task scheduling is based on an estimated cluster capacity detection and throttling algorithm.  
The goal is to maximize the number of created/scheduled task pods while minimizing the number of pods that 
cannot be scheduled to a node. The scheduler monitors the number of (task) pods the the k8s node scheduler 
cannot schedule and self-throttles accordingly. As a result, a _target capacity_ is dynamically determined
and constantly adjusted.

### Priority Escalation ###

A task's priority may be escalated (increased) when one of its dependencies is also in the
set of tasks ready to be started. The goal is to prevent lower priority dependencies
from impeding higher priority tasks.

Example:

- Task id=10 (kind=A) (priority=0)
- Task id=12 (kind=B) (priority=1) depends on: `A`

When scheduling both tasks, task(12) cannot run until task(10) has completed. This
condition effectively makes task(12) priority=0. To prevent this, the manager
will _escalate_ task(10) priority=1 to match task(12). 

### Macros ###

The manager supports injecting values into Addon and Extension specifications. 
Each macro has the syntax of: ${_name_}.  

Supported:

- ${**seq**:_pool_} - Number sequence generator. The _pool_ is the identifier and the beginning number.
  Example usage for network port assignment:
  ```yaml
   PORT_A: ${seq:8000}
   PORT_B: ${seq:8000}
  ```
  Results in:
  ```yaml
  PORT_A: 8000
  PORT_B: 8001
   ```

### Pods ###

Tasks are executed using Kubernetes Pods. When a task is _state=Ready_ to run, the
manager creates a Pod resource which is associated to the task. Task pods have the
following labels:
- app:`tackle`
- role:`task`
- task: _id_

The manager injects a few environment variables:

| Name         | Definition                                                                                                                                                                             |
|--------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| ADDON_HOME   | Path to an EmptyDir mounted as the working directory. Defined by the [HomeDir](https://github.com/konveyor/tackle2-hub/tree/main/settings#addon) setting.                              |
| SHARED_PATH  | Path to an EmptyDir mounted in all containers within the pod for sharing files. Defined by the [Shared.Path](https://github.com/konveyor/tackle2-hub/tree/main/settings#main) setting. |
| CACHE_PATH   | Path to a volume mounted in all containers in all pods for cached files. Defined by the [Cache.Path](https://github.com/konveyor/tackle2-hub/tree/main/settings#main) setting.         |
| HUB_BASE_URL | The hub API base url. Defined by the [Hub.URL](https://github.com/konveyor/tackle2-hub/tree/main/settings#addon) setting.                                                                                                                                  |
| TASK         | The task id (to be acted on).                                                                                                                                                          |
| TOKEN        | An authentication token for the hub API.                                                                                                                                               |

#### Retention ####

The pod associated with completed task is retained for a defined duration. After
which, the pod is deleted to prevent leaking pod resources indefinitely.

| State     | Retention (default)                                                                                |
|-----------|----------------------------------------------------------------------------------------------------|
| Succeeded | [Pod.Retention.Succeeded](https://github.com/konveyor/tackle2-hub/tree/main/settings#task-manager) |
| Failed    | [Pod.Retention.Failed](https://github.com/konveyor/tackle2-hub/tree/main/settings#task-manager)    |

#### Containers ####

The pod is created with a _main_ container (0) for the selected addon using the image 
defined by the Addon CR. Additional _sidecar_ containers are created for each extension
selected as defined by the Extension CR. After the _main_ (addon) container has terminated,
the manager will kill extension contains should they not terminate on their own. This is to
ensure complete termination of the pod after the addon container has terminated.

#### Log Collection ####

The manager _tails_ the log for each contain in the task pod. Each is stored as `File` in the
inventory and associated with the task as an attachment. The file is named using the
convention of the _container-name_.yaml.

## Task ##

Tasks are used to execute Addons.

### Properties ###

`*` indicates reported by addon.

| Name        | Definition                                                                                                                                      |
|-------------|-------------------------------------------------------------------------------------------------------------------------------------------------|
| ID          | Unique identifier.                                                                                                                              |
| CreateTime  | The timestamp of when the task was created.                                                                                                     |
| CreateUser  | The user (name) that created the task.                                                                                                          |
| UpdateUser  | The user (name) that last updated the task.                                                                                                     |
| Name        | The task mame (non-unique).                                                                                                                     |
| Kind        | The kind references a Task (kind) CR by name.                                                                                                   |
| Addon       | The addon to be executed. References an Addon CR by name. When not specified, the addon is selected based on kind.                              |
| Extension   | The list of extensions to be injected into the addon pod as _sidecar_ containers.                                                               |
| State       | The task state.  See: [States](https://github.com/konveyor/tackle2-hub/tree/main/task#states).                                                  |
| Locator     | The task locator. An arbitrary user-defined value used for lookup.                                                                              |
| Priority    | The task execution priority. See: [Priority](https://github.com/konveyor/tackle2-hub/tree/main/task#priority).                                  |
| Policy      | The task execution policy. Determines when task is postponed. See: [Policies](https://github.com/konveyor/tackle2-hub/tree/main/task#policies). | 
| TTL         | The task Time-To-Live in each state. See: [TTL](https://github.com/konveyor/tackle2-hub/tree/main/task#ttl-time-to-live).                       |
| Data        | The data provided to the addon. The schema is dictated by each addon. This may be _ANY_ document.                                               |
| Started     | The UTC timestamp when the task execution started.                                                                                              |
| Terminated  | The UTC timestamp when execution completed.                                                                                                     |
| Errors      | A list of reported errors. See: [Errors](https://github.com/konveyor/tackle2-hub/tree/main/task#errors).                                        |
| Events      | A list of reported task processing events. See: [Events](https://github.com/konveyor/tackle2-hub/tree/main/task#events).                        |
| Pod         | The fully qualified name of the pod created.                                                                                                    |
| Retries     | The number of times failure to create a pod is retried. This does not include when blocked by resource quota.                                   |
| Attached    | Files attached to the task.                                                                                                                     |
| \*Activity  | The activity (log) entries are reported by the addon. Intended to reflect what the addon is doing.                                              |
| \*Total     | Progress: The total number of items to be completed by the addon.                                                                               |
| \*Completed | Progress: The number of items completed by the addon.                                                                                           | |

### Events ###

Task events are used to record and report events related to task lifecycle and scheduling.

Fields:
- **Kind** - kind of event.
- **Count**: number of times the event is reported.
- **Reason** - The reason or cause of the event.
- **Last** - Timestamp when last reported.

| Event             | Meaning                                               |
|-------------------|-------------------------------------------------------|
| AddonSelected     | An addon has been selected.                           |
| ExtensionSelected | An extension has been selected.                       |
| ImageError        | The pod (k8s) reported an image error.                |
| PodNotFound       | The pod associated with a running pod does not exist. |
| PodCreated        | A pod has been created.                               |
| PodPending        | Pod (k8s) reported phase=Pending.                     |
| PodRunning        | The pod (k8s) reported phase=Running.                 |
| PodSucceeded      | The pod (k8s) has reported phase=Succeeded.           |
| PodFailed         | The pod (k8s) has reported phase=Error                |
| PodDeleted        | The pod has been deleted.                             |
| Escalated         | The manager has escalated the task priority.          |
| Released          | The task's resources have been released.              |
| ContainerKilled   | The specified (zombie) container needed to be killed. |

### Errors ###

Task errors are used to report problems with scheduling for execution.

Fields:
- **Severity** - Error severity. The values are at the discretion of the reporter.
- **Description** - Error description. Format: (_reporter_) _description_.

Note: A task may complete with a state=Succeeded with errors.

### States ###

`*` indicates _terminal_ states.

| State        | Definition                                                                                      |
|:-------------|:------------------------------------------------------------------------------------------------|
| Created      | The task has been created but not submitted.                                                    |
| Ready        | The task has been submitted to the manager and will be scheduled for execution.                 |
| Postponed    | The task has been postponed until another task has completed based on task scheduling _policy_. |
| QuotaBlocked | The task pod has been (temporarily) prevented from being created by k8s resource quota.         |
| Pending      | The task pod has been created and awaiting k8s scheduling.                                      |
| Running      | The task pod is running.                                                                        |
| \*Succeeded  | The task pod successfully completed.                                                            |
| \*Failed     | The task pod either failed to be started by k8s or completed with errors.                       |
| \*Canceled   | The task has been canceled.                                                                     |


### Policies ###

The task supports policies designed to influence scheduling.

| Name           | Definition                                               |
|----------------|----------------------------------------------------------|
| Isolated       | ALL other tasks are postponed while the task is running. |

### TTL (Time-To-Live) ###

The TTL determines how long at task may exist in a given state before the task
or associated resources are reaped.

## (Task) Kinds ###

The `Task` CR defines a name kind of task. Each kind may define:
- **Priority** - The default priority.
- **Dependencies** - List of dependencies (other task kinds). When created/ready concurrent,
  A task's dependencies must complete before the task is scheduled.
- **Metadata** - **TBD**.  

## Addons ##

An `Addon` CR defines a named addon (aka plugin). It defines functionality provide by an image to 
be executed. The definition includes a container specification and selection criteria. An addon 
may have extensions. See: [_Extensions_](https://github.com/konveyor/tackle2-hub/tree/main/task#extensions).

### Selection ###

When a task is created, either the `kind` or the `addon` may be specified. When the
`addon` is specified, the addon is selected by matching the name. When the `kind` is specified,
the addon is selected by matching the `Addon.Task` and evaluating the `Addon.Selector`.

Supported selector:
- tag:_category_=_tag_ - match application tags.
  ```yaml
  spec:
    task: ^(analyzer|tech-discovery)$
    selector: tag:Language=Java
- platform:_kind_=_kind_ - match platform kind.
  ```yaml
  spec:
    task: ^(analyzer|tech-discovery)$
    selector: platform:cloudfoundry
  ```

## Extensions ##

An extension defines an additional _sidecar_ container to be included in the task pod.

### Selection ###

When a task is created, it may define a list of extensions. When specified, addons are
selected by name. When not specified, addons are selected by matching the `Extension.Addon`
and evaluating the `Extension.Selector`. The selector includes logical `||` and `&&` operators
and `()` parens for grouping expressions.

Supported selector:
- tag:_category_=_tag_ - match application tags.
  ```yaml
  spec:
    addon: ^(analyzer|tech-discovery)$
    selector: tag:Language=Java
- platform:_kind_=_kind_ - match platform kind.
  ```yaml
  spec:
    addon: ^(analyzer|tech-discovery)$
    selector: platform:cloudfoundry
  ```

## Authorization ##

When the task pod is created and _Auth_ is enabled, a token is generated with the
necessary scopes. The token is mounted as a secret in the pod. The token is only 
valid while the task is running. 

## Reaping ###

A task may be reaped after existing in a state for the defined duration. 
This is to prevent orphaned or stuck tasks from leaking resources such as buckets and files.

| State     | Duration (default)                                                                          | Action   |
|-----------|---------------------------------------------------------------------------------------------|----------|
| Created   | [Reaper.Created](https://github.com/konveyor/tackle2-hub/tree/main/settings#task-manager)   | Deleted  |
| Succeeded | [Reaper.Succeeded](https://github.com/konveyor/tackle2-hub/tree/main/settings#task-manager) | Deleted  |
| Failed    | [Reaper.Failed](https://github.com/konveyor/tackle2-hub/tree/main/settings#task-manager)    | Released |

See [Reaper](https://github.com/konveyor/tackle2-hub/blob/main/reaper/README.md#reaper)
settings for details.

### Group ###

Task groups are used to create collections of tasks.

#### Modes ####

- Batch (default) - Used to create a _batch_ of un-ordered, homogeneous tasks. When the group is submitted, each member
  (task) is created. As part of task creation, group properties (kind, addon, extension, priority, policy, data) are 
  propagated to each member. The `Data` object is _merged_ into each Task.Data with Task.Data values taking precedent.
  Members (tasks) share group's bucket.
- Pipeline - Used to create a collection of heterogeneous tasks to be executed in order. When the group is submitted, 
  each member (task) is created and the FIRST task's state is set to Ready. No other properties are propagated. 
  As each task completes (state=Succeeded), the next task's state is set to Ready. 
  Members (tasks) share group's bucket.


| Name        | Definition                                                                                                         |
|-------------|--------------------------------------------------------------------------------------------------------------------|
| ID          | Unique identifier.                                                                                                 |
| CreateTime  | The timestamp of when the group was created.                                                                       |
| CreateUser  | The user (name) that created the task.                                                                             |
| UpdateUser  | The user (name) that last updated the task.                                                                        |
| Mode        | The task mode (Batch\|Pipeline).                                                                                   |
| Name        | The task mame (non-unique).                                                                                        |
| Kind        | The kind references a Task (kind) CR by name.                                                                      |
| Addon       | The addon to be executed. References an Addon CR by name. When not specified, the addon is selected based on kind. |
| Extension   | The list of extensions to be injected into the addon pod as _sidecar_ containers.                                  |
| State       | The task state.  See: [States](https://github.com/konveyor/tackle2-hub/tree/main/task#states).                                                                               |
| Priority    | The task execution priority. See: [Priority](https://github.com/konveyor/tackle2-hub/tree/main/task#priority).                                                                 |
| Policy      | The task execution policy. Determines when task is postponed. See: [Policies](https://github.com/konveyor/tackle2-hub/tree/main/task#policies).                                |
| Data        | The data provided to the addon. The schema is dictated by each addon. This may be _ANY_ document.                  |
| Tasks       | List of member tasks.                                                                                              |

#### Reaper ####

A task group may be reaped after existing in a state for the defined duration.
This is to prevent orphaned groups from leaking resources such as buckets and files.

| State     | Duration (default)    | Action   |
|-----------|-----------------------|----------|
| Created   | [Reaper.Created](https://github.com/konveyor/tackle2-hub/tree/main/settings#task-manager) | Deleted  |

See [Reaper](https://github.com/konveyor/tackle2-hub/blob/main/reaper/README.md#reaper)
settings for details.