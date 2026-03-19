
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

### How Reaping Works ###

The reaper runs periodically (controlled by the
[Frequency.Reaper](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#frequency)
setting) and processes resources in this order:

1. **TaskReaper** - Reaps tasks and releases their resources (buckets, files)
2. **GroupReaper** - Reaps task groups and releases their resources
3. **BucketReaper** - Identifies orphaned buckets and deletes expired ones
4. **FileReaper** - Identifies orphaned files and deletes expired ones

This order is important: tasks and groups must release their references first, allowing
buckets and files to be identified as orphans in the same iteration.

### Two-Phase Reaping for Referenced Resources ###

For buckets and files, reaping happens in two phases:

**Phase 1: Orphan Detection**
- The reaper scans all models that can reference the resource
- If no references are found, the resource is marked as orphaned by setting its `Expiration` field
- If references exist but the resource was previously marked as orphaned, the `Expiration` is cleared
  to account for short transient changes in ownership

**Phase 2: Expiration & Deletion**
- Resources with an expired `Expiration` timestamp are deleted
- Both the database record and filesystem artifact (for files/buckets) are removed

### Buckets ###

Buckets represent file trees stored in the filesystem. They may be referenced by:
- Application (BucketID field)
- Task (BucketID field)
- TaskGroup (BucketID field)

### Bucket Reaping Process ###

1. **Orphan Detection:** The BucketReaper scans all Applications, TaskGroups, and Tasks
   to find bucket references. Buckets not referenced are marked with an expiration timestamp.
2. **Grace Period:** Orphaned buckets have a grace period defined by the
   [Bucket.TTL](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#main)
   setting (default: 1 minute).
3. **Deletion:** When the expiration timestamp is reached, both the database record and
   the filesystem directory are deleted.
4. **Re-association:** If a bucket is re-referenced before expiration, the expiration
   timestamp is cleared and the bucket is preserved. This accounts for short transient
   changes in ownership where a resource briefly becomes orphaned during reassignment.

### Files ###

Files represent individual files stored in the filesystem. They may be referenced by:
- Task (Attached field - array of file IDs)
- TaskReport (Attached field - array of file IDs)
- Target (ImageID field)
- Rule (FileID field)
- AnalysisProfile (Files array field)

References are identified using the `ref` struct tag:
- `ref:"file"` - single file reference
- `ref:"[]file"` - array of file references

### File Reaping Process ###

1. **Orphan Detection:** The FileReaper uses the RefFinder to scan all models with
   `ref:"file"` or `ref:"[]file"` tags. Files not referenced are marked with an
   expiration timestamp.
2. **Grace Period:** Orphaned files have a grace period defined by the
   [File.TTL](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#main)
   setting (default: 12 hours).
3. **Deletion:** When the expiration timestamp is reached, both the database record and
   the filesystem file are deleted.
4. **Re-association:** If a file is re-referenced before expiration, the expiration
   timestamp is cleared and the file is preserved. This accounts for short transient
   changes in ownership where a resource briefly becomes orphaned during reassignment.

### Tasks ###

Tasks may be created then submitted in a _two-phase_ process. The reason is to support
a workflow that involves uploading files into the task's bucket. To do this, the user
must create the task, upload the files, then submit the task.

Tasks reference:
- Bucket
- File (via Attached field)

When a task is reaped, referenced objects such as buckets and files are _released_.
Releasing means removing the reference to allow them to be naturally reaped (garbage-collected).

### Task Reaping Rules by State ###

Tasks are reaped based on their state and TTL (time-to-live) configuration. Each task
can optionally define custom TTL values that override the default settings.

**Created** (not yet submitted):
- **With custom TTL:** Deleted after `TTL.Created` minutes from creation time
- **Without custom TTL:** Released (resources freed but task kept) after the
  [Reaper.Created](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#task-manager) setting
- **Pipeline protection:** Tasks in a pipeline group (Mode="Pipeline") are never reaped
  even if TTL expires, as they are waiting for their turn to execute

**Ready, Pending, QuotaBlocked** (waiting to execute):
- **With custom TTL:** Deleted after `TTL.Pending` minutes from creation time
- **Without custom TTL:** Never reaped (intentional - active tasks should complete naturally)

**Running** (actively executing):
- **With custom TTL:** Deleted after `TTL.Running` minutes from started time
- **Without custom TTL:** Never reaped (intentional - running tasks should complete naturally)

**Succeeded** (completed successfully):
- **With custom TTL:** Deleted after `TTL.Succeeded` minutes from termination time
- **Without custom TTL:** Released after the
  [Reaper.Succeeded](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#task-manager) setting

**Failed** (completed with error):
- **With custom TTL:** Deleted after `TTL.Failed` minutes from termination time
- **Without custom TTL:** Released after the
  [Reaper.Failed](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#task-manager) setting

**Canceled** (user-canceled):
- Currently not handled by the reaper

### Task Design Notes ###

- **Active states** (Ready, Pending, Running) are expected to complete or fail naturally.
  Only long-running or stuck tasks need explicit TTL protection.
- **Terminal states** (Succeeded, Failed) accumulate over time and need automatic cleanup
  via default settings.
- **TTL override:** Setting a custom TTL on a task overrides the default behavior for that
  specific task. Setting TTL to 0 falls back to the default settings.
- **Release vs Delete:** The Created state releases resources but keeps the task record
  when using default settings, while custom TTL results in deletion. This preserves task
  history for audit purposes.

### TaskGroups ###

TaskGroups represent collections of tasks that are executed together, either in parallel
or as a pipeline (sequential execution).

TaskGroups reference:
- Bucket
- Tasks (many-to-one relationship via Task.TaskGroupID)

### TaskGroup Reaping Rules by State ###

**Created** (not yet submitted):
- Deleted after a period defined by the
  [Reaper.Created](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#task-manager)
  setting (default: 72 hours)

**Ready** (submitted):
- **If group has no tasks:** Deleted after 1 hour
- **If group has tasks:** Bucket and List resources are released, but the group record
  is kept (tasks must be deleted first via TaskReaper)
- **Note:** Groups with tasks are never automatically deleted to preserve the relationship
  with existing tasks

### TaskGroup Design Notes ###

- TaskGroups in the Ready state are expected to complete their tasks, which are then
  individually reaped by the TaskReaper
- Once all tasks in a group are deleted, the group becomes eligible for deletion on the
  next reaper iteration (when `len(Tasks) == 0`)
- Groups release their bucket reference to allow the bucket to be garbage collected,
  while keeping the group record for audit and relationship integrity

## Configuration Reference ##

The reaper behavior is controlled by several environment variables and settings:

### Frequency

- **FREQUENCY_REAPER** - How often the reaper runs (default: 1 minute)
  - Controls the `Settings.Frequency.Reaper` value
  - Format: integer (minutes)

### Task Reaping TTLs

These settings define the default grace period before tasks are reaped when no custom TTL is set:

- **TASK_REAP_CREATED** - Created tasks (default: 4320 minutes / 72 hours)
- **TASK_REAP_SUCCEEDED** - Succeeded tasks (default: 4320 minutes / 72 hours)
- **TASK_REAP_FAILED** - Failed tasks (default: 43200 minutes / 30 days)

### Resource TTLs

- **BUCKET_TTL** - Orphaned buckets (default: 1 minute)
- **FILE_TTL** - Orphaned files (default: 720 minutes / 12 hours)

See [Settings Documentation](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md)
for complete configuration details.

**Note:** Pod retention and deletion is handled by the task manager, not the reaper.
See [Pod Retention settings](https://github.com/konveyor/tackle2-hub/blob/main/settings/README.md#task-manager)
for pod lifecycle management.

## Best Practices ##

### Using Custom TTLs

When creating tasks, you can override default reaping behavior by setting custom TTL values:

```json
{
  "name": "my-task",
  "addon": "analyzer",
  "ttl": {
    "created": 60,      // Delete if not submitted within 60 minutes
    "pending": 120,     // Delete if stuck in pending for 120 minutes
    "running": 240,     // Delete if running longer than 240 minutes
    "succeeded": 30,    // Delete 30 minutes after successful completion
    "failed": 1440      // Delete 1 day after failure
  }
}
```

**Recommendations:**
- Set `ttl.running` for long-running tasks to prevent resource leaks from stuck tasks
- Use longer `ttl.failed` values to preserve failed task data for debugging
- Set `ttl.created` for automated workflows to clean up abandoned tasks

### Audit Trail Preservation

- Created tasks without custom TTL are **released** (not deleted) to preserve audit history
- Terminal states (Succeeded/Failed) keep task records while releasing resources
- TaskGroups are kept as long as they have tasks to preserve relationships

### Pipeline Tasks

- Tasks in pipeline groups (Mode="Pipeline") are protected from reaping in Created state
- Even with expired TTL, pipeline tasks wait for their turn to execute
- Pipeline protection ensures sequential execution order is maintained

## Troubleshooting ##

### Tasks Not Being Reaped

1. Check if task is in a pipeline group (pipeline tasks in Created state are protected)
2. Verify task state - active states (Ready, Pending, Running) without TTL are never reaped
3. Check the task's custom TTL settings - they override defaults
4. Review reaper frequency setting - may need to wait for next iteration

### Resources Not Being Deleted

1. Check if resources are still referenced - use database queries to verify
2. Verify the expiration timestamp is set and has passed
3. Ensure filesystem permissions allow deletion
4. Check reaper logs for errors during deletion attempts

### Unexpected Deletions

1. Review custom TTL values on tasks - may be too aggressive
2. Check if tasks are being released when they shouldn't be
3. Verify bucket/file TTL settings - may be too short
4. Confirm no orphan detection false positives (check RefFinder logic)
