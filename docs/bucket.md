## Buckets ##
A bucket is a _container_ of files/directories stored in the hub.  The implementation (storage) is a 
directory which is owned by the Bucket REST resource.  Buckets are owned. Resources like Applications, 
Tasks, and TaskGroups will have references, modeled as subresources, to their bucket. If a bucket 
exists without an owner, it is an orphan and will be garbage collected by the Reaper. Otherwise, 
buckets follow conventional CRUD life-cycle patterns.

**GET** returns the Bucket resource.

**POST** Creates the Bucket resource and storage directory.

**DELETE** Deletes the Bucket resource and associated storage directory.


### Content ###
The Bucket itself has a _virtual_ content subresource that provides CRUD operations on the directory
tree managed by the bucket.

**GET** returns the file at the specified path as an octet-stream. When the path references a directory, it is 
packaged as a tarball and the `X-Directory=Expand` header set. A GET with an Accept=text/html header behaves like
a static file website.

**PUT** creates/updates the file/directory at the specified path. The intermediate directories are created 
as needed. When the file _isA_ directory, the client must upload a tarball and set the `X-Directory=Expand` header.
The body must be a miltipart form with a field named `file`.

**DELETE** deletes the file/directory at the specified path (but not intermediate directories).
