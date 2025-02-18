
## Manager ##

### Processing ###

The manager processes tasks (default: 1 second) in a _main_ loop. 
1. Fetch cluster resources using a cached k8s client.
2. Process queued task delete and cancel requests.
3. Delete orphaned pods. An orphaned pod is a pod within the namespace with task labels that does not
   correspond to a task in the running state.
4. Fetch running tasks; update their status based on associated pod status.
5. Kill zombies. Zombies are _sidecar_ containers that have not termineated on their own after the
   _main_ (addon) container has terminated.
6. Fetch and run new (state=Ready) tasks:
   1. select addon. See: _Addons.Selection_.
   2. select extensions. See: _Extensions.Selection_.
   3. create pod. See: _Pods_.

### Priority ###

Tasks with state=Ready are started based on their `Priority` property.
Priority=0 (default) is the lowest. Their is no maximum. This controls the order in which task
pods are created. Pod scheduling is still at the discretion of the node-scheduler. It is highly
recommended for administrators to create a k8s _Resource Quota_ to restrict the number of pods
to be scheduled at one time. Fewer pods maximizes the task priority influence over the execution order.

### Resource Quota ###

Kubernetes Resource Quotas are handled gracefully by the manager. When a pod cannot be created due
to quota restriction, the state=_QuotaBlocked_ and an event reported on the task. The manager will retry to create the pod in the
next processing cycle.

### Escalation ###

A task's priority may be escalated (increased) when an inversion created by Task (kind)
dependencies is detected.

Example:

- Task id=10 (kind=A) (priority=0)
- Task id=12 (kind=B) (priority=1) depends on: `A`

When scheduling both tasks, task(12) cannot run until task(10) has completed. This
condition effectively makes task(12) priority=0. To prevent this, the manager
will _escalate_ task(10) priority=1 to match task(12). The goal is to prevent dependencies
from impeding higher priority tasks.

### Preemption ###

To prevent priority inversions, the manager supports preempting a _Running_ task so that a
higher priority _Ready_ task pod may be scheduled. Preemption is the act of killing (deleting)
the pod of a _running_ task, so that the higher _blocked_ task may be created/scheduled
by the node-scheduler. A task is considered _blocked_ when it cannot be created due to
a resource quota (state=QuotaBlocked) or cannot be scheduled by the node-scheduler
(state=Pending) for a defined duration (default: 1 minute).
To trigger preemption, the _blocked_ task must have Policy.PreemptEnabled=TRUE. When
the need for preemption is detected, the manger will preempt a percentage (default: 10%) of the 
newest, lower priority tasks processing cycle. To prevent _thrashing_ a preempted task will 
be postponed for a defined duration (default: 1 minute).
When a task is preempted:
1. The pod is deleted.
2. The task state is reset to Ready.
3. A `Preempted` event is recorded.

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

Tasks are executed using Kubernetes Pods. When a task is _Ready_ to run, the
manager creates a Pod resource which is associated to the task.

#### Retention ####

The pod associated with completed task is retained for a defined duration. After
which, the pod is deleted to prevent leaking pod resources indefinitely.

| State     | Retention (default) |
|-----------|---------------------|
| Succeeded | 1 (second)          |
| Failed    | 72 (hour)           |

#### Containers ####

The pod is created with a _main_ container (0) for the selected addon using the image 
defined by the Addon CR. Additional _sidecar_ containers are created for each extension
selected as defined by the Extension CR. After the _main_ (addon) container has terminated,
the manager will kill extension contains should they not terminate on their own. This is to
ensure complete termination of the pod.

#### Log Collection ####

The manager _tails_ the log for each contain in the task pod. Each is stored as `File` in the
inventory and associated with the task as an attachment. The file is named using the
convention of the container _name_.yaml.

## Task ##

Tasks are used to execute Addons.

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
| Preempted         | The task has been preempted by the manager.           |
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

| State        | Definition                                                                                    |
|:-------------|:----------------------------------------------------------------------------------------------|
| Created      | The task has been created but not submitted.                                                  |
| Ready        | The task has been submitted to the manager and will be scheduled for execution.               |
| Postponed    | The task has been postponed until another task has completed based on task scheduling _policy_. |
| QuotaBlocked | The task pod has been (temporarily) prevented from being created by k8s resource quota.       |
| Pending      | The task pod has been created and awaiting k8s scheduling.                                    |
| Running      | The task pod is running.                                                                      |
| \*Succeeded  | The task pod successfully completed.                                                          |
| \*Failed     | The task pod either failed to be started by k8s or completed with errors.                     |
| \*Canceled   | The task has been canceled.                                                                   |


### Policies ###

The task supports policies designed to influence scheduling.

| Name           | Definition                                               |
|----------------|----------------------------------------------------------|
| Isolated       | ALL other tasks are postponed while the task is running. |
| PreemptEnabled | When (true), the task _may_ trigger preemption.          |
| PreemptExempt  | When (true), the task may NOT be preempted.              |


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
may have extensions. See: _Extensions_.

### Selection ###

When a task is created, either the `kind` or the `addon` may be specified. When the
`addon` is specified, the addon is selected by matching the name. When the `kind` is specified,
the addon is selected by matching the `Addon.Task` and evaluating the `Addon.Selector`.

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
  selector: tag:Language=Java
  ```

## Authorization ##

When the task pod is created and _Auth_ is enabled, a token is generated with the
necessary scopes. The token is mounted as a secret in the pod. The token is only 
valid while the task is running. 

## Reaping ###

A task may be reaped after existing in a state for defined duration. This is to prevent
orphaned or stuck tasks from leaking resources such as buckets and files.

| State     | Duration (default) | Action   |
|-----------|--------------------|----------|
| Created   | 72 (hour)          | Deleted  |
| Succeeded | 72 (hour)          | Deleted  |
| Failed    | 30 (day)           | Released |

