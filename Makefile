GOBIN ?= ${GOPATH}/bin

PKG = ./addon/... \
      ./api/... \
      ./auth/... \
      ./cmd/... \
      ./encryption/... \
      ./importer/... \
      ./k8s/... \
      ./model/... \
      ./settings/... \
      ./volume/... \
      ./controller/... \
      ./task/... 

BUILD = --tags json1 -o bin/hub github.com/konveyor/tackle2-hub/cmd

# Build ALL commands.
cmd: hub addon

# Run go fmt against code
fmt:
	go fmt ${PKG}

# Run go vet against code
vet:
	go vet ${PKG}

# Build hub
hub: generate fmt vet
	go build ${BUILD}

# Build manager binary with compiler optimizations disabled
debug: generate fmt vet
	go build -gcflags=all="-N -l" ${BUILD}

docker: vet
	go build ${BUILD}

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	go run ./cmd/main.go

# Generate manifests e.g. CRD, Webhooks
manifests: controller-gen
	${CONTROLLER_GEN} ${CRD_OPTIONS} \
		crd rbac:roleName=manager-role \
		paths="./..." output:crd:artifacts:config=generated/crd/bases output:crd:dir=generated/crd

# Generate code
generate: controller-gen
	${CONTROLLER_GEN} object:headerFile="./generated/boilerplate" paths="./..."

# Find or download controller-gen.
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.5.0 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# Build SAMPLE ADDON
addon: fmt vet
	go build -o bin/addon github.com/konveyor/tackle2-hub/hack/cmd/addon

# Build Swagger API spec into ./docs directory
docs-swagger:
	swag init -g api/base.go
