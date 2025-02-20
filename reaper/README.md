
## Reaper ##

The hub inventory contains a set of entities that are garbage collected (reaped) 
rather than deleted through cascading. These entities may be referenced by multiple
types of entities so managing deletion though cascade delete can be complex. Instead,
the relation between objects is modeled as _references_. When the reference count
is zero(0), it is considered an _orphan_ and the referenced object is assigned 
an expiration TTL (time-to-live). Orphans are not deleted immediately to account 
for transient conditions. If not re-associated when the TTL has expired,
the orphan is deleted.

### Buckets ###

Buckets represent file trees and may be referenced by:
- applications
- tasks
- task groups

TTL (default: 1 minute).

### Files ###

Files represent files and may be referenced by:
- applications
- tasks
- task-reports
- targets
- rules

TTL (default: 1 minute).

### Tasks ###


