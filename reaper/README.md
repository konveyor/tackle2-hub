
## Reaper ##

The hub inventory contains a set of entities that are garbage collected (reaped) 
rather than deleted through cascading. These entities may be referenced by multiple
types of entities so managing deletion though cascade delete can be complex. Instead,
the relation between objects is modeled as _references_. When the reference count
is zero(0), it is considered an _orphan_ and the referenced object is assigned 
an expiration TTL (time-to-live). This _grace-period_ is designed to account for 
transient conditions such as a transfer of ownership. If not re-associated when the 
TTL has expired, the orphan is deleted.

There are also non-referenced entities that need to be reaped.

### Buckets ###

Buckets represent file trees and may be referenced by:
- Application
- Task
- TaskGroup

The TTL after orphaned is defined by the
[Bucket.TTL](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#main) setting.

### Files ###

Files represent files and may be referenced by:
- Task
- TaskReport
- Target
- Rule

The TTL after orphaned is defined by the
[File.TTL](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#main) setting.

### Tasks ###

Tasks may be created then submitted in a _two-phase_ process. The reason is to support
a workflow that involves uploading files into the task's bucket. To do this, the user
must create the task, upload the files, then submit the task. To prevent leaking tasks
that are created and never submitted, the reaper will delete tasks that have not been
submitted after a period defined by the
[Reaper.Created](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#task-manager) setting.

Submitted tasks are never reaped. As a result, there is a need for objects
referenced by tasks to be reaped. Tasks reference:
- Bucket
- File
- Pod

Reaped objects such as buckets and files are said to be _released_ when the task is
reaped. Releasing simply means to remove the reference to allow them to be
naturally reaped (garbage-collected).

**Succeeded** tasks:
- Released after a period defined by the
  [Reaper.Succeeded](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#task-manager) setting.
- Pod deleted after a period defined by the
  [Pod.Renention.Succeeded](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#task-manager) setting.

**Failed** tasks:
- Released after a period defined by the
  [Reaper.Failed](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#task-manager) setting.
- Pod deleted after a period defined by the
  [Pod.Retention.Failed](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#task-manager) setting.
