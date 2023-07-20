### Reaping ###

#### Buckets and Files ####
Buckets and Files are garbage collected resources.
They are considered orphans when no longer referenced by another resource. When an orphan is
detected, it is assigned an expiration (date/time). The expiration is a kind of _grace period_
intended to support two-phase assignment to an owner or re-assignment.
