Addons

An addon is an extension of the Hub and consists of a container (image) and a
manifest (Custom Resource). The CR defines the addon and registers it with the Hub.
It is intended to provide business logic that contributes to the application 
*knowledge-base* (inventory) or produces modernization assets. An addon is essentially
just a binary written in any language that is integrated through the Hub API. The addon
must report its status using the task (addon) report:
1. Report addon started: `POST /tasks/:id/report` Task.State=Running.
2. Report addon activity & progress: `PUT /tasks/:id/report` (repeat as needed).
3. Report addon completed: `PUT /tasks/:id/report` Task.State=(Succeeded|Failed)

The knowledge about an application consists of:
- Application (model).
- Source repository.
- Tags.
- Application Facts (key/value pair).
- Files (stored in the application *bucket*).
- Assessments.

Modernization assets may be stored in:
- Source repository.
- Application Facts (key/value pair).
- Files (stored in the application *bucket*).

## Definition

An Addon is defined and registered with the Hub using a Custom Resource (CR).

* **image** - The addon image.
* **imagePullPolicy** - An (optional) image pull policy. Defaults to `IfNotPresent`.
* **resources** - An *(optional)* standard k8s pod container [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) specification.
* **mounts** - An optional list of volumes to be mounted in the addon pod. **DEPRECATED**
    * **claim** - PVC name.
    * **name** - The name of the directory in `/mnt` to mount the volume.

Example:
```
apiVersion: tackle.konveyor.io/v1alpha1
kind: Addon
metadata:
  name: admin
  namespace: konveyor-tackle
spec:
  image: quay.io/konveyor/tackle2-addon:latest
  mounts:
  - claim: tackle-maven-volume-claim
    name: m2
  resources:
    requests:
      cpu: 50m
      memory: 50Mi
```

## Runtime (pod)

### Environment

The pod is started with the following environment variables:
  - **ADDON_WORKING_DIR** - Path to addon working directory.
  - **HUB_BASE_URL** - Base URL for the hub API.
  - **TOKEN** - An authentication token.
  - **TASK** - An associated task ID.

The bucket volume is mounted in `/buckets`.

The addon container must have the default entry point set.

### Go Binding

The addon is expected to integrate with the Hub inventory using the
Hub REST API. For convenience, the Hub (package)provides 
a `Go` [binding](https://github.com/konveyor/tackle2-hub/tree/main/docs/binding.txt) that simplifies integration.

#### Run API.

The Run() method simplifies/wraps addon execution.
Provides addon (task) reporting of:
 - Addon started.
 - Addon succeeded when no error is returned.
 - Addon failed when error is returned.
 - Addon failed when panic detected.

Recommended addon _main_.

```
package main

import (
    hub "github.com/konveyor/tackle2-hub/addon"
    "github.com/konveyor/tackle2-hub/api"
)

var (
    // hub integration. 
    addon = hub.Addon
    Log   = hub.Log
)

type SoftError = hub.SoftError

//
// main
func main() {
    addon.Run(func() (err error) {
        // 
        // Get the addon data associated with the task. 
        d := &Data{}
        _ = addon.DataWith(d)
        if err != nil {
            return
        }
        // 
        // EXECUTE ADDON HERE.
    }
}
```

### Running your addon locally.

An addon can be _run_ locally either on the command line or in an IDE.  An addon is just a
binary/image that integrates with the Tackle Hub API. When using the Go binding (recommended)
a few environment variables need to be set.  The `Task.Data` must contain the addon's input _document_.
The `Task.Application` may reference an application ID. In effect, you are simulating what
the Hub task manager is doing.

Steps:
 - Be sure tackle is installed with auth disabled.
 - Ensure the `/buckets` directory exists.
 - Using the tackle UI, create an application (optional).
 - Using cURL, create a task. Example in hack/add/task.sh. Be sure **not** to submit the task.
 - Set environment variables:
   - **ADDON_WORKING_DIR** - An (optional) working directory. Defaults to `/tmp`
   - **HUB_BASE_URL** - Tackle _ingress-ip_/hub. For example: `$ export HUB_BASE_URL=http://192.168.49.2/hub`
   - **TASK** - The ID of the task you created.
 - Execute your addon (either).
   - `$ make run`
   - Launch in your IDE.
 - Monitor the task progress using cURL or your browser at: `http://$HUB_BASE_URL/tasks/$TASK`.

The addon may be run against the same task as many times as needed.

