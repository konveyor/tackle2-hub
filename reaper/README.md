
## Reaper ##

The hub inventory contains a set of entities that are garbage collected (reaped) 
rather than deleted through cascading. These entities may be referenced by multiple
types of entities so managing deletion though cascade delete can be complex. Instead,
the relation between objects is modeled as _references_. When the reference count
is zero(0), it is considered an _orphan_ and the referenced object is assigned 
an expiration TTL (time-to-live). Orphans are not deleted immediately to account 
for transient conditions. If not re-associated when the TTL has expired,
the orphan is deleted.

There are also non-referenced entities that need to be reaped.

### Buckets ###

Buckets represent file trees and may be referenced by:
- Application
- Task
- TaskGroup

TTL after orphaned (default: 1 minute).

### Files ###

Files represent files and may be referenced by:
- Task
- TaskReport
- Target
- Rule

TTL after orphaned (default: 1 minute).

### Tasks ###

Tasks may be created then submitted in a _two-phase_ process. The reason is to support
a workflow that involves uploading files into the task's bucket. To do this, the user
must create the task, upload the files, then submit the task. To prevent leaking tasks
that are created and never submitted, the reaper will delete tasks that have not been
submitted after a defined period (default: 72 hours).

Submitted tasks are never reaped. As a result, there is a need for objects
referenced by tasks to be reaped. Tasks reference:
- bucket
- file
- pod

Reaped objects such as buckets and files are said to be _released_ when the task is
reaped. Releasing simply means to remove the reference to allow them to be
naturally reaped (garbage-collected).

**Succeeded** tasks:
- Released after a defined period (default: 72 hours).
- Pod deleted after a defined period (default: 1 minute)

**Failed** tasks:
- Released after a defined period (default: 30 days).
- Pod deleted after a defined period (default 72 hours).
