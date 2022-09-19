Addons

#Definition

An Addon is defined by a Custom Resource (CR):
* **image** - The addon image.
* **task** - Task (mode):
    * **command** - An *(optional)* standard k8s pod command specificaion.
    * **resources** - An *(optional)* standard k8s pod resources specificaion.
    * * **mounts** - An optional list of volumes to be mounted in the addon pod.
        * **claim** - PVC name.
        * **name** - The name of the directory in `/mnt` to mount the volume.
* **service** - Service (mode):
*   * **command** - An *(optional)* standard k8s pod command specificaion.
    * **resources** - An *(optional)* standard k8s pod resources specificaion.
    * * **mounts** - An optional list of volumes to be mounted in the addon pod.
        * **claim** - PVC name.
        * **name** - The name of the directory in `/mnt` to mount the volume.
    * **rbac** - An *(optional)* specification of roles and scopes.
        * **role** - A role name.
        * **scopes** - A list of scopes included in the role.
* **frontend** - An (optional) micro-frontend definition.
    * **navigation** - Defines how the ui is integrated.
    * **bundle** - The bundle URL.
    


3.x Example (long term):
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

# Runtime (pod)

The Addon is launched by the Hub in one of two modes: *Task* & *Service*.  Theimage is expected to have an entry point (command) for each mode. For *service* mode, the hub will create a Service and Pod as defined. The Pod will be started with a secret containing the URL and token used to communicate with the Hub. For task mode, the hub will create a Pod as defined. The Pod will be started with a secret containing the URL and token used to communicate with the Hub.
