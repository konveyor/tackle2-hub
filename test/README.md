# Hub tests

## REST API

```
$ export HUB_ENDPOINT="http://$(minikube ip)/hub"   # Or e.g. http://localhost:8080
$ go test -v ./test/api/...
```