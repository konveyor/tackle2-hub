# Hub Tests

Hub tests consist of following parts:
- Unit tests
- REST API tests
- WIP (more or less) Integration tests
- WIP Export/import tests

All tests can be executed with ```$ make test``` which will run all available tests.

## General information

- Tests are written in golang to fit well to the Konveyor project components.
- Each test is responsible for setup its test data and clean it when finished.
- The main way of interacting with Hub is its API, to make testing easier, following tools are provided:
  - Test ```client``` package wrapping addon.Client provides API methods like Get/Post/etc., Should/Must and other equality assertions.
  - Hub's ```API``` package provides predefined routes and resources struct definition.
  - Each resource has a test package that tests the resource endpoint methods and provides fixtures and basic CRUD methods (like ```application.Create(&testApp)```) to other tests.


## REST API

API tests can be executed on locally running Hub in development mode (without other Konveyor components using DISCONNECTED=1).

```
$ export DISCONNECTED=1
$ make run
```

```
$ export HUB_ENDPOINT=http://localhost:8080
$ go test -v ./test/api/
```

Or tests can run against running Hub installation (example with minikube below).

```
$ export HUB_ENDPOINT="http://$(minikube ip)/hub"
$ export KEYCLOAK_ADMIN_USER="admin"
$ export KEYCLOAK_ADMIN_PASS=""
$ go test -v ./test/api/
```

Sample output
```
$ make test-api 
echo "Using Hub API from http://192.168.39.236/hub"
Using Hub API from http://192.168.39.236/hub
go test -v ./test/api/...
{"level":"info","ts":1679056454.0379033,"logger":"addon","msg":"Addon (adapter) created."}
{"level":"info","ts":1679056454.041117,"logger":"addon","msg":"|201|  POST /hub/auth/login"}
...
=== RUN   TestApplicationUpdateName
=== RUN   TestApplicationUpdateName/Update_application_Pathfinder
{"level":"info","ts":1679056454.1980724,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.2010725,"logger":"addon","msg":"|204|  PUT /hub/applications/1"}
{"level":"info","ts":1679056454.2033865,"logger":"addon","msg":"|200|  GET /hub/applications/1"}
{"level":"info","ts":1679056454.2071495,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
=== RUN   TestApplicationUpdateName/Update_application_Minimal_application
{"level":"info","ts":1679056454.2095335,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.2123275,"logger":"addon","msg":"|204|  PUT /hub/applications/1"}
{"level":"info","ts":1679056454.2146227,"logger":"addon","msg":"|200|  GET /hub/applications/1"}
{"level":"info","ts":1679056454.2210302,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
--- PASS: TestApplicationUpdateName (0.03s)
    --- PASS: TestApplicationUpdateName/Update_application_Pathfinder (0.01s)
    --- PASS: TestApplicationUpdateName/Update_application_Minimal_application (0.01s)
PASS
ok  	github.com/konveyor/tackle2-hub/test/api/application	0.194s
```
