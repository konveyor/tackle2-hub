## Buckets ##
A bucket is _container_ of files/directories sorted in the hub.  The implementation (storage) is a directory which is 
owned by the Bucket REST resource.  Buckets follow the conventional CRUD life-cycle as other resources but with 
some notable exceptions. Buckets have the concept of bucket ownership.  Buckets are designed to provide file storage for other resources and cannot
exist on their own.  A bucket without an owner (parent) is an orphan and subject to garbage collection by the reaper.
Bucket owners have references to their buckets which are modeled in the API as subresources.
Current bucket owners:
- Applications
- TaskGroup
- Task

A **GET** will return the `Bucket`.

A **POST** Create the `Bucket` and storage directory.

A **DELETE** will delete the `Bucket` and associated storage directory.


### Content ###
The Bucket itself has a _virtual_ content subresource that provides GET/PUT/DELETE operations.  
For example: a route of `/buckets/1/this/that/thing`  `this/that/thing` is the path to the `thing` subresource (file) 
within the bucket.

A **GET** will return `thing` as an octet-stream. when thing is a directory, it is packaged as a tarball and
the `X-Directory=Expand` header set. Bucket (content) may be inspected using a Web Browser. A **GET** on 
a Bucket route with Accept=text/html will return an index.html.

A **PUT** will (upload) store `thing` in the bucket. Intermediate directories are automatically created. When 
`thing` is a directory, the client must package thing as a tarball and set the `X-Directory=Expand` header.

A **DELETE** will delete `thing` (but not intermediate directories).

## Files ##

A file represents a single file stored in the hub. Files follow the conventional CRUD life-cycle as other resources but with
some notable exceptions. Files have the concept of file ownership.  Files are designed to provide file storage for other resources and cannot
exist on their own.  A file without an owner (parent) is an orphan and subject to garbage collection by the reaper.
Current bucket owners:
- RuleSet
- Rule

A **GET** will return the file based on the Accept header:
- Accept=applications/json: The `File` resource.
- Accept=(other): The file (content) as an octet-stream.

A **PUT** Create the `File` resource and store the file (content).

A **DELETE** will delete the `File` resource and associated file (content).

### Reaping ###
Buckets and Files are garbage collected resources.
They are considered orphans when no longer referenced by another resource. When an orphan is
detected, it is assigned an expiration (date/time). The expiration is a kind of _grace period_
intended to support two-phase assignment to an owner or re-assignment.

### Streaming ###

Both Bucket and File content is uploaded using a multipart form.  Field: `File`=octet-stream.
