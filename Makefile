GOBIN ?= ${GOPATH}/bin
IMG   ?= tackle2-hub:latest

PKG = ./addon/... \
      ./api/... \
      ./auth/... \
      ./cmd/... \
      ./database/... \
      ./encryption/... \
      ./importer/... \
      ./k8s/... \
      ./migration/... \
      ./model/... \
      ./settings/... \
      ./controller/... \
      ./task/...  \
      ./tracker/...

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

# Build image
docker-build:
	docker build -t ${IMG} .

podman-build:
	podman build -t ${IMG} .
	
# Build manager binary with compiler optimizations disabled
debug: generate fmt vet
	go build -gcflags=all="-N -l" ${BUILD}

docker: vet
	go build ${BUILD}

# Run against the configured Kubernetes cluster in ~/.kube/config
run: fmt vet
	go run ./cmd/main.go

run-addon:
	go run ./hack/cmd/addon/main.go

# Generate manifests e.g. CRD, Webhooks
manifests: controller-gen
	controller-gen ${CRD_OPTIONS} \
		crd rbac:roleName=manager-role \
		paths="./..." output:crd:artifacts:config=generated/crd/bases output:crd:dir=generated/crd

# Generate code
generate: controller-gen
	controller-gen object:headerFile="./generated/boilerplate" paths="./..."

# Find or download controller-gen.
controller-gen:
	if [ "$(shell which controller-gen)" = "" ]; then \
	  set -e ;\
	  CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	  cd $$CONTROLLER_GEN_TMP_DIR ;\
	  go mod init tmp ;\
	  go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.10.0 ;\
	  rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	fi ;\

# Build SAMPLE ADDON
addon: fmt vet
	go build -o bin/addon github.com/konveyor/tackle2-hub/hack/cmd/addon

# Build Swagger API spec into ./docs directory
docs-swagger:
	${GOBIN}/swag init -g api/base.go

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
