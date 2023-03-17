# Hub tests

## REST API


Execute with

```
$ export HUB_ENDPOINT="http://$(minikube ip)/hub"   # Or e.g. http://localhost:8080
$ go test -v ./test/api/...
```

Sample output
```
$ make test-api 
echo "Using Hub API from http://192.168.39.236/hub"
Using Hub API from http://192.168.39.236/hub
go test -v ./test/api/...
{"level":"info","ts":1679056454.0379033,"logger":"addon","msg":"Addon (adapter) created."}
{"level":"info","ts":1679056454.041117,"logger":"addon","msg":"|201|  POST /hub/auth/login"}
=== RUN   TestApplicationBucket
{"level":"info","ts":1679056454.0440674,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.0486953,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
--- PASS: TestApplicationBucket (0.01s)
=== RUN   TestApplicationCreate
=== RUN   TestApplicationCreate/Create_application_Pathfinder
{"level":"info","ts":1679056454.0515816,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.0564034,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
=== RUN   TestApplicationCreate/Create_application_Minimal_application
{"level":"info","ts":1679056454.0588925,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.0630262,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
--- PASS: TestApplicationCreate (0.01s)
    --- PASS: TestApplicationCreate/Create_application_Pathfinder (0.01s)
    --- PASS: TestApplicationCreate/Create_application_Minimal_application (0.01s)
=== RUN   TestApplicationNotCreateDuplicates
{"level":"info","ts":1679056454.0662534,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.0690432,"logger":"addon","msg":"|409|  POST /hub/applications"}
{"level":"info","ts":1679056454.073798,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
--- PASS: TestApplicationNotCreateDuplicates (0.01s)
=== RUN   TestApplicationNotCreateWithoutName
{"level":"info","ts":1679056454.0759547,"logger":"addon","msg":"|400|  POST /hub/applications"}
--- PASS: TestApplicationNotCreateWithoutName (0.00s)
=== RUN   TestApplicationDelete
=== RUN   TestApplicationDelete/Delete_application_Pathfinder
{"level":"info","ts":1679056454.0787778,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.0827587,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
{"level":"info","ts":1679056454.0850236,"logger":"addon","msg":"|404|  GET /hub/applications/1"}
=== RUN   TestApplicationDelete/Delete_application_Minimal_application
{"level":"info","ts":1679056454.087594,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.091441,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
{"level":"info","ts":1679056454.0933266,"logger":"addon","msg":"|404|  GET /hub/applications/1"}
--- PASS: TestApplicationDelete (0.02s)
    --- PASS: TestApplicationDelete/Delete_application_Pathfinder (0.01s)
    --- PASS: TestApplicationDelete/Delete_application_Minimal_application (0.01s)
=== RUN   TestApplicationFactCRUD
{"level":"info","ts":1679056454.0960307,"logger":"addon","msg":"|201|  POST /hub/applications"}
=== RUN   TestApplicationFactCRUD/Fact_pet_application_Pathfinder
{"level":"info","ts":1679056454.0983863,"logger":"addon","msg":"|201|  POST /hub/applications/1/facts/pet"}
{"level":"info","ts":1679056454.1002402,"logger":"addon","msg":"|200|  GET /hub/applications/1/facts/pet"}
{"level":"info","ts":1679056454.105619,"logger":"addon","msg":"|204|  PUT /hub/applications/1/facts/pet"}
{"level":"info","ts":1679056454.1076562,"logger":"addon","msg":"|200|  GET /hub/applications/1/facts/pet"}
{"level":"info","ts":1679056454.1095133,"logger":"addon","msg":"|204|  DELETE /hub/applications/1/facts/pet"}
{"level":"info","ts":1679056454.1113322,"logger":"addon","msg":"|404|  GET /hub/applications/1/facts/pet"}
=== RUN   TestApplicationFactCRUD/Fact_address_application_Pathfinder
{"level":"info","ts":1679056454.113441,"logger":"addon","msg":"|201|  POST /hub/applications/1/facts/address"}
{"level":"info","ts":1679056454.1153314,"logger":"addon","msg":"|200|  GET /hub/applications/1/facts/address"}
{"level":"info","ts":1679056454.1174152,"logger":"addon","msg":"|204|  PUT /hub/applications/1/facts/address"}
{"level":"info","ts":1679056454.1193273,"logger":"addon","msg":"|200|  GET /hub/applications/1/facts/address"}
{"level":"info","ts":1679056454.1211658,"logger":"addon","msg":"|204|  DELETE /hub/applications/1/facts/address"}
{"level":"info","ts":1679056454.123087,"logger":"addon","msg":"|404|  GET /hub/applications/1/facts/address"}
{"level":"info","ts":1679056454.126788,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
--- PASS: TestApplicationFactCRUD (0.03s)
    --- PASS: TestApplicationFactCRUD/Fact_pet_application_Pathfinder (0.02s)
    --- PASS: TestApplicationFactCRUD/Fact_address_application_Pathfinder (0.01s)
=== RUN   TestApplicationFactsList
{"level":"info","ts":1679056454.1293018,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.1314142,"logger":"addon","msg":"|201|  POST /hub/applications/1/facts/pet"}
{"level":"info","ts":1679056454.1336236,"logger":"addon","msg":"|201|  POST /hub/applications/1/facts/address"}
=== RUN   TestApplicationFactsList/Fact_list_application_Pathfinder_with_facts
{"level":"info","ts":1679056454.135608,"logger":"addon","msg":"|200|  GET /hub/applications/1/facts"}
=== RUN   TestApplicationFactsList/Fact_list_application_Pathfinder_with_facts/
{"level":"info","ts":1679056454.1377947,"logger":"addon","msg":"|200|  GET /hub/applications/1/facts"}
{"level":"info","ts":1679056454.1421986,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
--- PASS: TestApplicationFactsList (0.02s)
    --- PASS: TestApplicationFactsList/Fact_list_application_Pathfinder_with_facts (0.00s)
    --- PASS: TestApplicationFactsList/Fact_list_application_Pathfinder_with_facts/ (0.00s)
=== RUN   TestApplicationGet
=== RUN   TestApplicationGet/Get_application_Pathfinder
{"level":"info","ts":1679056454.14488,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.1474292,"logger":"addon","msg":"|200|  GET /hub/applications/1"}
{"level":"info","ts":1679056454.1515367,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
=== RUN   TestApplicationGet/Get_application_Minimal_application
{"level":"info","ts":1679056454.1545959,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.1569495,"logger":"addon","msg":"|200|  GET /hub/applications/1"}
{"level":"info","ts":1679056454.1612723,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
--- PASS: TestApplicationGet (0.02s)
    --- PASS: TestApplicationGet/Get_application_Pathfinder (0.01s)
    --- PASS: TestApplicationGet/Get_application_Minimal_application (0.01s)
=== RUN   TestApplicationList
{"level":"info","ts":1679056454.164118,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.1663258,"logger":"addon","msg":"|201|  POST /hub/applications"}
{"level":"info","ts":1679056454.169194,"logger":"addon","msg":"|200|  GET /hub/applications"}
{"level":"info","ts":1679056454.191005,"logger":"addon","msg":"|204|  DELETE /hub/applications/1"}
{"level":"info","ts":1679056454.1954293,"logger":"addon","msg":"|204|  DELETE /hub/applications/2"}
--- PASS: TestApplicationList (0.03s)
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
