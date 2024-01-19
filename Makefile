GOBIN ?= $(GOPATH)/bin
GOIMPORTS = $(GOBIN)/goimports
CONTROLLERGEN = $(GOBIN)/controller-gen
IMG   ?= tackle2-hub:latest
HUB_BASE_URL ?= http://localhost:8080

PKG = ./addon/... \
      ./api/... \
      ./assessment/... \
      ./auth/... \
      ./binding/... \
      ./controller/... \
      ./cmd/... \
      ./database/... \
      ./encryption/... \
      ./importer/... \
      ./k8s/... \
      ./metrics/... \
      ./migration/... \
      ./model/... \
      ./nas/... \
      ./reaper/... \
      ./seed/... \
      ./settings/... \
      ./tar/... \
      ./task/...  \
      ./test/...  \
      ./tracker/...

PKGDIR = $(subst /...,,$(PKG))

BUILD = --tags json1 -o bin/hub github.com/konveyor/tackle2-hub/cmd

# Build ALL commands.
cmd: hub addon

# Format the code.
fmt: $(GOIMPORTS)
	goimports -w $(PKGDIR)

# Run go vet against code
vet:
	go vet $(PKG)

# Build hub
hub: generate fmt vet
	go build $(BUILD)

# Build image
docker-build:
	docker build -t $(IMG) .

podman-build:
	podman build -t $(IMG) .
	
# Build manager binary with compiler optimizations disabled
debug: generate fmt vet
	go build -gcflags=all="-N -l" $(BUILD)

docker: vet
	go build $(BUILD)

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	go run ./cmd/main.go

run-addon:
	go run ./hack/cmd/addon/main.go

# Generate manifests e.g. CRD, Webhooks
manifests: $(CONTROLLERGEN)
	$(CONTROLLERGEN) $(CRD_OPTIONS) \
		crd rbac:roleName=manager-role \
		paths="./..." output:crd:artifacts:config=generated/crd/bases output:crd:dir=generated/crd

# Generate code
generate: $(CONTROLLERGEN)
	$(CONTROLLERGEN) object:headerFile="./generated/boilerplate" paths="./..."

# Ensure controller-gen installed.
$(CONTROLLERGEN):
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.10.0

# Ensure goimports installed.
$(GOIMPORTS):
	go install golang.org/x/tools/cmd/goimports@latest

# Build SAMPLE ADDON
addon: fmt vet
	go build -o bin/addon github.com/konveyor/tackle2-hub/hack/cmd/addon

docs: docs-html docs-openapi3 docs-binding

# Build Swagger API spec into ./docs directory
docs-swagger:
	$(GOBIN)/swag init -g pkg.go --dir api,assessment

# Build OpenAPI 3.0 docs
docs-openapi3: docs-swagger
	curl -X POST -H "Content-Type: application/json" -d @docs/swagger.json https://converter.swagger.io/api/convert | jq > docs/openapi3.json

# Build HTML docs from Swagger API spec
docs-html: docs-openapi3
	redoc-cli bundle -o docs/index.html docs/openapi3.json

# Build binding doc.
docs-binding:
	go doc --all addon > docs/binding.txt

.PHONY: start-minikube
START_MINIKUBE_SH = ./bin/start-minikube.sh
start-minikube:
ifeq (,$(wildcard $(START_MINIKUBE_SH)))
	@( \
	set -e ;\
	mkdir -p $(dir $(START_MINIKUBE_SH)) ;\
	curl -sSLo $(START_MINIKUBE_SH) https://raw.githubusercontent.com/konveyor/tackle2-operator/main/hack/start-minikube.sh ;\
	chmod +x $(START_MINIKUBE_SH) ;\
	)
endif
	$(START_MINIKUBE_SH);

.PHONY: install-tackle
INSTALL_TACKLE_SH = ./bin/install-tackle.sh
install-tackle:
ifeq (,$(wildcard $(INSTALL_TACKLE_SH)))
	@( \
	set -e ;\
	mkdir -p $(dir $(INSTALL_TACKLE_SH)) ;\
	curl -sSLo $(INSTALL_TACKLE_SH) https://raw.githubusercontent.com/konveyor/tackle2-operator/main/hack/install-tackle.sh ;\
	chmod +x $(INSTALL_TACKLE_SH) ;\
	)
endif
	$(INSTALL_TACKLE_SH);

# Run test targets always (not producing test dirs there).
.PHONY: test test-api test-integration migration

# Run unit tests (all tests outside /test directory).
test:
	go test -count=1 -v $(shell go list ./... | grep -v "hub/test")

# Run Hub REST API tests.
test-api:
	HUB_BASE_URL=$(HUB_BASE_URL) go test -count=1 -p=1 -v ./test/api/...

# Run Hub test suite.
test-all: test-unit test-api

migration:
	hack/next-migration.sh

