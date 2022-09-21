Addons

An addon is an extension of the Hub and consists of a container (image) and a
manifest (Custom Resource). The CR defines the addon and registers it with the Hub.
It is intended to provide business logic that contributes to the application 
*knowledge-base* (inventory) or produces modernization artifacts. 

The knowledge about an application consists of:
- Application (model).
- Source or binary repository.
- Tags.
- Facts (key/value pair).
- Files (stored in the application *bucket*).
- Assessments.

Modernization artifacts may be stored in:
- Source or binary repository.
- Facts (key/value pair).
- Files (stored in the application *bucket*).

Addons optionally support the following deployment modes:
- **Task** - The asynchronous mode intended for long-running and/or high resource usage. Each request is run in a separate pod on-demand.
- **Service** - The synchronous (interactive) mode intended for read-only or short running actions requested though the Addon's REST API. 

# Definition

An Addon is defined and registered with the Hub using a Custom Resource (CR).

### Current (Tackle 2)

* **image** - The addon image.
* **resources** - An *(optional)* standard k8s pod container [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) specification.
* **mounts** - An optional list of volumes to be mounted in the addon pod.
    * **claim** - PVC name.
    * **name** - The name of the directory in `/mnt` to mount the volume.

2.x Example (short term):
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

### Future (Tackle 3)

* **image** - The addon image.
* **task** - Task (mode):
    * **command** - An *(optional)* standard k8s pod command specificaion.
    * **resources** - An *(optional)* standard k8s pod container [resources](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers) specification.
    * * **mounts** - An optional list of volumes to be mounted in the addon pod.
        * **claim** - PVC name.
        * **name** - The name of the directory in `/mnt` to mount the volume.
* **service** - Service (mode):
  * **command** - An *(optional)* standard k8s pod command specificaion.
  * **resources** - An *(optional)* standard k8s pod resources specificaion.
  * **mounts** - An optional list of volumes to be mounted in the addon pod.
     * **claim** - PVC name.
     * **name** - The name of the directory in `/mnt` to mount the volume.
* **rbac** - An *(optional)* specification of roles and scopes.
    * **role** - A role name.
    * **scopes** - A list of scopes included in the role.
* **frontend** - An (optional) micro-frontend definition.
    * **navigation** - Defines how the ui is integrated.
    * **bundle** - The bundle URL.

Example:
```
apiVersion: tackle.konveyor.io/v1alpha1
kind: Addon
metadata:
  name: admin
  namespace: konveyor-tackle
spec:
  image: quay.io/konveyor/tackle2-addon:latest
  task:
      command: /usr/bin/addon
      mounts:
      - claim: tackle-maven-volume-claim
        name: m2
      resources:
        requests:
          cpu: 50m
          memory: 50Mi
  service:
      command: /usr/bin/addon-service
      resources:
        requests:
          cpu: 50m
          memory: 50Mi
      api:
          rbac:
          - role: admin
            scopes:
             - widgets.post
             - widgets.get
             - widgets.put
             - widgets.delete
          - role: migrator
            scopes:
             - widgets.get
          - role: develper
            scopes:
             - widgets.get
  frontend:
      navigation: application.inventory.Widgets
      bundle: https://github.com/myui.json
```

# Runtime (pod)

### Current (Tackle 2)

The Addon is deployed by the Hub in one of (or both) two modes: *Task* & *Service*. 
In both modes, the addon container (image) is deployed in a Pod with the following
resources mounted:
- Hub secret `/tmp/secret.json` which includes:
  - **Hub**:
    - **Token** - Hub (auth) API token.
    - **Application** - An *(optional)* application ID.
    - **Task** - An *(optional)* Task ID.
    - **Variant** - An *(optional)* Task variant.
    - **Encryption**:
      - **Passphrase**: Encryption key passphrase.
  - **Addon**: Addon *data* passed through the task.
- Volumes defined in the Addon CR.

### Future (Tackle 3)

Notes:
- *The API will render decrypted data based on token*.
- *The addon will GET the task as needed by ID*.
- *The secret fields may be stored as ENVARs.*

The Addon is deployed by the Hub in one of (or both) two modes: *Task* & *Service*.
In both modes, the addon container (image) is deployed in a Pod with the following
resources mounted:
- Hub secret `/tmp/secret.json` which includes:
  - **Token** - Hub (auth) API token.  `$TOKEN`
  - **Application** - An *(optional)* application ID. `$APPLICATION`
  - **Task** - An *(optional)* Task ID. `$TASK`
- Volumes defined in the Addon CR.