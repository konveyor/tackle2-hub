
## Manager ##

### Processing ###

The manager processes tasks (each second) in a _main_ loop. 
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



#### Retention ####

#### Containers ####

#### Log Collection ####

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


