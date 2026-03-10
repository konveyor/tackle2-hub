# Task Scheduler Specification

This document defines the functional requirements for the task scheduler. Tests should verify these behaviors, not the implementation details.

---

## Task Lifecycle

### Behavior: Task State Machine

**Given**: A task is created in the database
**When**: The scheduler processes the task through its lifecycle
**Then**: Task transitions follow this state machine:
- `Created` → `Ready` (when submitted)
- `Ready` → `Pending` (when pod is created)
- `Pending` → `Running` (when pod starts)
- `Running` → `Succeeded` (when addon container exits 0)
- `Running` → `Failed` (when any container exits non-zero, except 137)
- `Running` → `Ready` (when container exits 137 and retries < max retries)
- Any state → `Canceled` (when user cancels)

**Invalid Transitions**:
- `Succeeded` → any state (terminal)
- Direct `Running` → `Ready` without going through `Failed` (except exit 137)

**Test**: TestTaskLifecycle

---

### Behavior: Task Retry on Container Kill

**Given**: A running task's container exits with code 137 (killed)
**When**: Retries < Settings.Hub.Task.Retries
**Then**:
- Task state transitions to `Ready`
- Pod is deleted
- Retries counter increments
- Task can be rescheduled

**When**: Retries >= Settings.Hub.Task.Retries
**Then**: Task transitions to `Failed`

**Test**: TestTaskRetryOnKill

---

### Behavior: Task with Missing Pod (Zombie Detection)

**Given**: Task with state `Pending` or `Running` but pod doesn't exist
**When**: Scheduler's updateRunning phase runs
**Then**:
- Task transitions to `Ready`
- Task.Pod field cleared
- Event `PodNotFound` recorded
- Started and Terminated timestamps cleared

**Test**: TestTaskMissingPod

---

## Scheduler Loop Behaviors

### Behavior: Orphaned Pod Cleanup

**Given**: A pod exists with task label but no corresponding task in database
**When**: Scheduler's deleteOrphanPods phase runs
**Then**:
- Pod is deleted
- Event logged: "Orphan pod found"

**Test**: TestOrphanPodCleanup

---

### Behavior: Retained Pod Cleanup

**Given**: Task with state `Succeeded` or `Failed` and Retained=true
**When**: Current time - Terminated > retention period
**Then**:
- Pod is deleted
- Task.Retained set to false
- Task remains in database

**Retention periods**:
- Succeeded: Settings.Hub.Task.Pod.Retention.Succeeded
- Failed: Settings.Hub.Task.Pod.Retention.Failed

**Test**: TestRetainedPodCleanup

---

### Behavior: Zombie Pod Cleanup

**Given**: Task with state `Succeeded` or `Failed` but pod still running
**When**: ContainerKilled event exists for > 1 minute
**Then**: Pod is forcibly deleted

**Test**: TestZombiePodCleanup

---

## Scheduling Rules

### Behavior: RuleUnique - Concurrent Task Limiting

**Given**: Application 1 has 3 tasks with Kind=analyzer, State=Ready
**When**: Scheduler runs startReady
**Then**:
- First 2 tasks transition to Pending (pods created)
- Third task remains Ready with state=Postponed
- Event recorded: "Rule:Unique matched:3, other:2"

**When**: One of the first two tasks completes
**Then**:
- Third task becomes eligible
- Third task transitions to Pending on next cycle

**Matching criteria**:
- Same Application/Platform (subject)
- Same Kind (when specified)
- Same Addon (fallback)

**Test**: TestRuleUnique

---

### Behavior: RuleDeps - Task Dependencies

**Given**:
- Task kind "analysis" has dependency on kind "discovery"
- Task 1: Kind=discovery, State=Running, ApplicationID=1
- Task 2: Kind=analysis, State=Ready, ApplicationID=1

**When**: Scheduler runs postpone rules
**Then**:
- Task 2 remains Ready with state=Postponed
- Event recorded: "Rule:Dependency matched:2, other:1"

**When**: Task 1 completes
**Then**: Task 2 becomes eligible on next cycle

**Test**: TestRuleDeps

---

### Behavior: RuleIsolated - Isolated Task Policy

**Given**:
- Task 1: Policy.Isolated=true, State=Running
- Task 2: Policy.Isolated=true, State=Ready

**When**: Scheduler runs postpone rules
**Then**:
- Task 2 remains Ready with state=Postponed
- Event recorded: "Rule:Isolated matched:2, other:1"

**When**: Task 1 completes
**Then**: Task 2 becomes eligible

**Test**: TestRuleIsolated

---

## Priority Escalation

### Behavior: Dependency Priority Escalation

**Given**:
- Task 1: Kind=discovery, Priority=5, State=Ready, ApplicationID=1
- Task 2: Kind=analysis, Priority=10, State=Ready, ApplicationID=1, depends on discovery

**When**: Scheduler runs adjustPriority
**Then**:
- Task 1 priority escalated to 10
- Event recorded: "Escalated:1, by:2"

**When**: Task 1 state is Pending
**Then**:
- Pod is deleted
- Task 1 transitions back to Ready
- Task 1 will be rescheduled with new priority

**Test**: TestPriorityEscalation

---

## Addon and Extension Selection

### Behavior: Addon Selection by Tag

**Given**:
- Application tagged with "Language=Java"
- Addon "analyzer" with selector: "tag:Language=Java"
- Task: Kind=analyzer, ApplicationID=1

**When**: Scheduler selects addon for task
**Then**:
- Task.Addon = "analyzer"
- Event recorded: AddonSelected

**Test**: TestAddonSelectionByTag

---

### Behavior: Extension Selection by Tag

**Given**:
- Application tagged with "Language=Java"
- Extension "java" with selector: "tag:Language=Java"
- Addon "analyzer" selected
- Extension "java" has addon selector matching "analyzer"

**When**: Scheduler selects extensions for task
**Then**:
- Task.Extensions contains "java"
- Event recorded: ExtSelected

**Test**: TestExtensionSelectionByTag

---

### Behavior: Complex Selector Expression

**Given**: Application with tags "Language=Java", "Framework=Spring"
**When**: Addon selector is "tag:Language=Java && tag:Framework=Spring"
**Then**: Addon matches

**When**: Addon selector is "tag:Language=Java && !tag:Language=Python"
**Then**: Addon matches

**When**: Addon selector is "tag:Language=Python || tag:Language=Java"
**Then**: Addon matches

**Test**: TestComplexSelectors

---

## Capacity Management

### Behavior: Capacity Estimation

**Given**: Cluster has scheduled 5 task pods, 0 unscheduled
**When**: CapacityMonitor adjusts capacity
**Then**: Capacity increases by growth rate (default 1.05x)

**Given**: Cluster has 3 scheduled pods, 2 unscheduled (PodReasonUnschedulable)
**When**: CapacityMonitor adjusts capacity
**Then**: Capacity reduced to 1 (3 scheduled - 2 newly unscheduled)

**Test**: TestCapacityEstimation

---

### Behavior: Capacity Exceeded Pauses Scheduling

**Given**: CapacityMonitor detects unscheduled pods (unscheduled > 0)
**When**: Scheduler runs startReady
**Then**:
- No new pods created
- Log message: "Capacity exceeded: pod creation paused."

**Test**: TestCapacityExceeded

---

## Quota Enforcement

### Behavior: Pod Quota Limits

**Given**:
- Pod quota = 10
- Current task pods = 9
- 5 ready tasks

**When**: Scheduler runs startReady
**Then**:
- Only 1 pod created (quota allows 10 - 9 = 1)
- Remaining tasks stay Ready

**Test**: TestPodQuotaLimit

---

### Behavior: Quota Blocked State

**Given**: Task pod quota exhausted
**When**: Task attempts to create pod
**Then**:
- Task state = QuotaBlocked
- Event recorded: QuotaBlocked with reason
- Task becomes eligible when quota available

**Test**: TestQuotaBlocked

---

## Pod Construction

### Behavior: Extension Environment Variable Propagation

**Given**:
- Extension "java" with env vars: PORT=8000, MAVEN_OPTS=-Dmaven.repo.local=/cache/m2
- Task uses extension "java"

**When**: Pod is created
**Then**: Addon container has environment variables:
- `_EXT_JAVA_PORT=8000`
- `_EXT_JAVA_MAVEN_OPTS=-Dmaven.repo.local=/cache/m2`

**Test**: TestExtensionEnvPropagation (exists: TestScheduler)

---

### Behavior: Standard Volume Mounts

**Given**: Any task pod
**When**: Pod is created
**Then**: All containers have volume mounts:
- Name: `addon`, MountPath: Settings.Addon.HomeDir
- Name: `shared`, MountPath: Settings.Addon.SharedDir
- Name: `cache`, MountPath: Settings.Addon.CacheDir

**Test**: TestStandardVolumeMounts (exists: TestScheduler)

---

### Behavior: Task Secret with JWT Token

**Given**: Task with ID=123, Addon="analyzer"
**When**: Pod is created
**Then**:
- Secret created with name prefix "task-123-"
- Secret contains key "HUB_TOKEN" with JWT
- JWT claims include: user="addon:analyzer", task=123
- Environment variable HUB_TOKEN references secret

**Test**: TestTaskSecret

---

## Pipeline Mode

### Behavior: Pipeline Task Sequencing

**Given**: TaskGroup with Mode=Pipeline containing tasks [T1, T2, T3]
**When**: TaskGroup is submitted
**Then**:
- T1.State = Ready
- T2.State = Created
- T3.State = Created

**When**: T1 completes with state=Succeeded
**Then**: T2.State = Ready

**When**: T2 completes with state=Succeeded
**Then**: T3.State = Ready

**Test**: TestPipelineMode

---

### Behavior: Pipeline Failure Cancellation

**Given**: TaskGroup with Mode=Pipeline containing tasks [T1, T2, T3]
**When**: T1 State=Running, T2 State=Created, T3 State=Created
**And**: T1 completes with state=Failed
**Then**:
- T2.State = Canceled
- T3.State = Canceled
- Event recorded: "Canceled:2, when (pipelined) task:1 failed."

**Test**: TestPipelineFailure

---

## Batch Mode

### Behavior: Batch Task Execution

**Given**: TaskGroup with Mode=Batch containing tasks [T1, T2, T3]
**When**: TaskGroup is submitted
**Then**:
- T1.State = Ready
- T2.State = Ready
- T3.State = Ready
- All tasks can run concurrently (subject to other rules)

**Test**: TestBatchMode

---

## Error Handling

### Behavior: Image Pull Error

**Given**: Pod container in Waiting state with reason containing "invalid", "error", "never", or "cannot"
**When**: Scheduler updates running tasks
**Then**:
- Task.State = Failed
- Event recorded: ImageError
- Error recorded with container name and reason

**Test**: TestImagePullError

---

### Behavior: Unschedulable Pod

**Given**: Pod with condition PodScheduled and reason=PodReasonUnschedulable
**When**: Scheduler updates running tasks
**Then**:
- Event recorded: PodUnschedulable with message
- Task remains in Pending state

**Test**: TestUnschedulablePod

---

### Behavior: Pod Snapshot on Completion

**Given**: Task completes (Succeeded or Failed)
**When**: Scheduler updates running tasks
**Then**:
- File created: "pod.yaml" containing:
  - Pod YAML specification
  - Pod events formatted as table
- File attached to task

**Test**: TestPodSnapshot

---

### Behavior: Container Termination on Retention

**Given**: Task completes and pod retention > 0
**When**: Pod has running containers
**Then**:
- Command executed in each running container: "sh -c kill 1"
- Event recorded: ContainerKilled
- Pod marked as Retained=true

**Test**: TestContainerTermination

---

## Task Priority Scheduling

### Behavior: Priority-Based Ordering

**Given**: Tasks with priorities [1, 5, 10, 3] all Ready
**When**: Scheduler creates pods
**Then**: Tasks start in priority order: 10, 5, 3, 1

**Given**: Tasks with same priority [5, 5, 5] and IDs [10, 20, 30]
**When**: Scheduler creates pods
**Then**: Tasks start in ID order: 10, 20, 30 (older first)

**Test**: TestPriorityOrdering

---

## Macro Injection

### Behavior: Sequence Macro Injection

**Given**: Container with env var PORT="${seq:8000}"
**When**: Injector processes container
**Then**: First reference → PORT=8000

**Given**: Container with:
- Args: ["--port", "${seq:8000}", "--admin-port", "${seq:8000}"]

**When**: Injector processes container
**Then**:
- Args: ["--port", "8000", "--admin-port", "8001"]

**Test**: TestSeqInjector

---

## Manager API

### Behavior: Create Task with State Validation

**Given**: Request to create task with State="Running"
**When**: Manager.Create() called
**Then**: Error returned: "state must be (Created|Ready)"

**Given**: Request to create task with State="" (empty)
**When**: Manager.Create() called
**Then**: Task created with State="Created"

**Test**: TestCreateTaskStateValidation

---

### Behavior: Update Task State Restrictions

**Given**: Task with State=Created
**When**: Update request changes Name, Kind, Addon, State
**Then**: All fields updated

**Given**: Task with State=Running
**When**: Update request changes Name, Priority
**Then**: Changes discarded (no update)

**Given**: Task with State=Ready
**When**: Update request changes Name, Locator, Policy, TTL
**Then**: Only these fields updated

**Test**: TestUpdateTaskRestrictions

---

### Behavior: Cancel Task

**Given**: Task with State=Running and pod exists
**When**: Manager.Cancel() called
**Then**:
- Pod deleted
- Task.State = Canceled
- Event recorded: Canceled
- Task.Bucket cleared

**Given**: Task with State=Succeeded
**When**: Manager.Cancel() called
**Then**: No action (discarded)

**Test**: TestCancelTask

---

### Behavior: Delete Task

**Given**: Task exists with pod
**When**: Manager.Delete() called
**Then**:
- Pod deleted
- Task deleted from database

**Test**: TestDeleteTask

---

## Notes for Test Implementation

### Test Environment Configuration

All tests should use:
- In-memory SQLite database
- K8s simulator with configurable pod lifecycle timing
- Settings.Frequency.Task = 100ms (fast test execution)
- Pre-seeded resources:
  - TagCategory "Language"
  - Tag "Java"
  - Addon "analyzer"
  - Extension "java"

### Common Test Patterns

**Setup**:
```go
tc := setup(g)
defer tc.teardown()
app := tc.seed(g)
```

**Create Task**:
```go
task := &model.Task{
    Name:          "test-task",
    Kind:          "analyzer",
    State:         impTask.Ready,
    ApplicationID: &app.ID,
}
err := tc.DB.Create(task).Error
g.Expect(err).To(gomega.BeNil())
```

**Start Manager**:
```go
tc.newManager(g)
```

**Wait for Completion**:
```go
time.Sleep(2500 * time.Millisecond)
```

**Verify State**:
```go
var updated model.Task
err := tc.DB.First(&updated, task.ID).Error
g.Expect(err).To(gomega.BeNil())
g.Expect(updated.State).To(gomega.Equal(impTask.Succeeded))
```

---

## Future Enhancements

The following behaviors are anticipated but not yet implemented:

- **Task Preemption**: Higher priority tasks evict lower priority tasks
- **Resource Limits**: CPU/Memory quota enforcement
- **Multi-Cluster Scheduling**: Distribute tasks across clusters
- **Task Groups with Dependencies**: DAG-based task execution
- **Custom Retry Policies**: Per-task retry configuration
- **Webhook Notifications**: Task state change notifications

---

**Last Updated**: 2026-03-10
