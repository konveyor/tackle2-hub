## Test (reference) addon.

The addon will list files in /etc and create artifacts for each file.
Adds a fact: `Listed=true`.
Adds a tag (type): `DIRECTORY`
Adds tags: `LISTED`, `TEST` and associates with the application.

To run locally:

#### Environment variables:
- **HUB_BASE_URL** - The hub API base URL. Default: `localhost:8080`.
- **TASK** - The associated task ID.

#### Steps:
- Set environment variables.
- Edit hack/add/task.sh as needed to reference an application ID and run. This only needs to be done once.
- Run the addon in your IDE or using `make run-addon`.

#### Example:
```
$ minikube ip
192.168.49.2
$ export HOST=192.168.49.2/hub
$ export HUB_BASE_URL=http://$HOST
$ hack/add/task.sh
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
100   424  100   281  100   143  11758   5983 --:--:-- --:--:-- --:--:-- 18434
{
  "id": 16,
  "createUser": "admin.noauth",
  "updateUser": "",
  "createTime": "2022-12-02T14:01:53.69882779Z",
  "name": "Test",
  "locator": "app.1.test",
  "addon": "test",
  "data": {
    "path": "/etc"
  },
  "application": {
    "id": 1,
    "name": ""
  },
  "state": "Created",
  "bucket": "/buckets/cfcf0899-73bc-43e3-95f9-a41a0c3388fb"
}
$ export TASK=16
$ make run-addon
```
