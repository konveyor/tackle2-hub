## Files ##

A file represents a single file stored in the hub. Files are owned. Resources like RuleSets,
and Rules will have references, modeled as subresources, to a File. If a bucket exists without an owner, 
it is an orphan and will be garbage collected by the Reaper. Otherwise, files follow conventional CRUD 
life-cycle patterns.

**GET** returns the file based on the Accept header:
- Accept=applications/json: The File resource.
- Accept=(other): The file (content) as an octet-stream.

**PUT** creates/updates the File resource and stores the file (content).
The body must be a miltipart form with a field named `file`.

**DELETE** deletes the File resource and associated file (content).
