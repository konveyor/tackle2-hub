## Buckets ##
A bucket is a _container_ of files/directories stored in the hub.  The implementation (storage) is a 
directory which is owned by the Bucket REST resource.  Buckets are owned. Resources like Applications, 
Tasks, and TaskGroups will have references, modeled as subresources, to their bucket. If a bucket 
exists without an owner, it is an orphan and will be garbage collected by the Reaper. Otherwise, 
buckets follow conventional CRUD life-cycle patterns.

**GET** returns the `Bucket` resource.

**POST** Creates the `Bucket` resource and storage directory.

**DELETE** Deletes the `Bucket` resource and associated storage directory.


### Content ###
The Bucket itself has a _virtual_ content subresource that provides GET/PUT/DELETE operations. For example: a route 
of `/buckets/1/this/that/thing`, `this/that/thing` is the path to the `thing` subresource (file) within the bucket.

**GET** returns `thing` as an octet-stream. when `thing` is a directory, it is packaged as a tarball and
the `X-Directory=Expand` header set. Bucket (content) may be inspected using a Web Browser. A **GET** on 
a Bucket route with Accept=text/html returns a directory index.

**PUT** stores `thing` in the bucket. Intermediate directories are automatically created. When 
`thing` is a directory, the client must package it as a tarball and set the `X-Directory=Expand` header.
The body must be a miltipart form with a field named `file`.

**DELETE** deletes `thing` (but not intermediate directories).

