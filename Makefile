GOPATH ?= $(HOME)/go
GOBIN ?= $(GOPATH)/bin
GOIMPORTS = $(GOBIN)/goimports
GOSWAG = $(GOBIN)/swag
CONTROLLERGEN = $(GOBIN)/controller-gen
IMG   ?= tackle2-hub:latest
HUB_BASE_URL ?= http://localhost:8080
FRONTEND_AUTH_DIR = $(CURDIR)/internal/frontend/auth/content
FRONTEND_AUTH_MARKER = $(FRONTEND_AUTH_DIR)/.dist-built

PKG = ./internal/... \
      ./shared/... \
      ./cmd/...

PKGDIR = $(subst /...,,$(PKG))

BUILD = --tags json1 -o bin/hub github.com/konveyor/tackle2-hub/cmd

# Build ALL commands.
cmd: hub shared/addon

# Format the code.
fmt: $(GOIMPORTS)
	$(GOIMPORTS) -w $(PKGDIR)

# Run go vet against code
vet: frontend-stub
	go vet $(PKG)

# Build hub
hub: frontend generate fmt vet
	go build $(BUILD)

# Build image
docker-build:
	docker build -t $(IMG) .

podman-build:
	podman build -t $(IMG) .

# Build manager binary with compiler optimizations disabled
debug: frontend generate fmt vet
	go build -gcflags=all="-N -l" $(BUILD)

docker: vet
	go build $(BUILD)

# Run against the configured Kubernetes cluster in ~/.kube/config
run: frontend fmt vet
	go run ./cmd/main.go

# Build frontend
.PHONY: frontend-stub
frontend-stub:
	mkdir -p $(FRONTEND_AUTH_DIR)/dist
	touch $(FRONTEND_AUTH_DIR)/dist/empty

.PHONY: frontend
frontend: $(FRONTEND_AUTH_MARKER)

$(FRONTEND_AUTH_DIR)/node_modules: $(FRONTEND_AUTH_DIR)/package-lock.json $(FRONTEND_AUTH_DIR)/package.json
	cd $(FRONTEND_AUTH_DIR) && npm ci
	@touch $@

$(FRONTEND_AUTH_MARKER): $(FRONTEND_AUTH_DIR)/node_modules $(shell find $(FRONTEND_AUTH_DIR)/src $(FRONTEND_AUTH_DIR)/branding $(FRONTEND_AUTH_DIR)/*.ts -type f 2>/dev/null)
	cd $(FRONTEND_AUTH_DIR) && npm run build
	@touch $@

# Clean all frontend builds
.PHONY: clean-frontend
clean-frontend:
	cd $(FRONTEND_AUTH_DIR) && npm run clean
	rm -f $(FRONTEND_AUTH_MARKER)

.PHONY: login
login: bin/login

bin/login: ./hack/cmd/login/main.go
	go build -o bin/login ./hack/cmd/login/main.go

run-addon:
	go run ./hack/cmd/addon/main.go

# Generate manifests e.g. CRD, Webhooks
manifests: $(CONTROLLERGEN)
	$(CONTROLLERGEN) $(CRD_OPTIONS) \
		crd rbac:roleName=manager-role \
		paths="./..." output:crd:artifacts:config=internal/generated/crd/bases output:crd:dir=internal/generated/crd

# Generate code
generate: $(CONTROLLERGEN)
	$(CONTROLLERGEN) object:headerFile="./internal/generated/boilerplate" paths="./..."

# Ensure controller-gen installed.
$(CONTROLLERGEN):
	go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.17.0

# Ensure goimports installed.
$(GOIMPORTS):
	go install golang.org/x/tools/cmd/goimports@v0.24

# Ensure swag installed.
$(GOSWAG):
	go install github.com/swaggo/swag/cmd/swag@latest

# Build SAMPLE ADDON
addon: fmt vet
	go build -o bin/addon github.com/konveyor/tackle2-hub/hack/cmd/addon

docs: docs-html docs-openapi3 docs-binding

# Build Swagger API spec into ./docs directory
docs-swagger: $(GOSWAG)
	$(GOSWAG) init --parseDependency --parseInternal --parseDepth 1 -g pkg.go --dir internal/api,internal/assessment

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
	@{ \
	set -e ;\
	mkdir -p $(dir $(START_MINIKUBE_SH)) ;\
	curl -sSLo $(START_MINIKUBE_SH) https://raw.githubusercontent.com/konveyor/tackle2-operator/main/hack/start-minikube.sh ;\
	chmod +x $(START_MINIKUBE_SH) ;\
	}
endif
	$(START_MINIKUBE_SH);

.PHONY: install-tackle
INSTALL_TACKLE_SH = ./bin/install-tackle.sh
install-tackle:
ifeq (,$(wildcard $(INSTALL_TACKLE_SH)))
	@{ \
	set -e ;\
	mkdir -p $(dir $(INSTALL_TACKLE_SH)) ;\
	curl -sSLo $(INSTALL_TACKLE_SH) https://raw.githubusercontent.com/konveyor/tackle2-operator/main/hack/install-tackle.sh ;\
	chmod +x $(INSTALL_TACKLE_SH) ;\
	}
endif
	$(INSTALL_TACKLE_SH);

# Run test targets always (not producing test dirs there).
.PHONY: test test-api test-integration migration test-binding test-auth test-all

# Run unit tests (all tests outside /test directory).
test: frontend-stub
	go test -count=1 -v $(shell go list ./... | grep -v "hub/test")

test-db: frontend-stub
	go test -count=1 -timeout=6h -v ./database...

# Run Hub REST API tests.
test-api:
	# Deprecated

test-binding: frontend-stub
	for pkg in $$(go list ./test/binding/...); do \
	  HUB_BASE_URL="$(HUB_BASE_URL)" go test -count=1 -v -failfast -parallel=1 "$$pkg" || exit 1; \
	done

test-auth: frontend-stub
	for pkg in $$(go list ./test/auth/...); do \
	  HUB_BASE_URL="$(HUB_BASE_URL)" go test -count=1 -v -failfast "$$pkg" || exit 1; \
	done

# Run Hub test suite.
test-all: test test-binding test-auth

migration:
	hack/next-migration.sh

