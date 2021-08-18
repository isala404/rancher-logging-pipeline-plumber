# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:crdVersions=v1"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif


# Dependencies
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen

KUSTOMIZE = $(shell pwd)/bin/kustomize

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /bin/bash -o pipefail
.SHELLFLAGS = -ec

all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

manifests: setup ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	cd pkg/sdk && $(CONTROLLER_GEN) $(CRD_OPTIONS) paths="./..." output:crd:artifacts:config=../../config/crd/bases
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role paths="./controllers/..." output:rbac:artifacts:config=./config/rbac

update-chart: manifests ## Build helm chart with the manager.
	cp config/crd/bases/* ./charts/logging-pipeline-plumber/crds/
	sed -i 's/- service_account.yaml/#- service_account.yaml/' ./config/rbac/kustomization.yaml
	$(KUSTOMIZE) build config/rbac > ./charts/logging-pipeline-plumber/templates/role.yaml
	sed -i 's/#- service_account.yaml/- service_account.yaml/' ./config/rbac/kustomization.yaml
	sed -i 's/controller-manager/{{ include "logging-pipeline-plumber.serviceAccountName" . }}/' ./charts/logging-pipeline-plumber/templates/role.yaml
	sed -i 's/manager-rolebinding/{{ include "logging-pipeline-plumber.fullname" . }}/' ./charts/logging-pipeline-plumber/templates/role.yaml
	sed -i 's/manager-role/{{ include "logging-pipeline-plumber.fullname" . }}/' ./charts/logging-pipeline-plumber/templates/role.yaml
	sed -i 's/namespace: system/namespace: {{ .Release.Namespace }}/' ./charts/logging-pipeline-plumber/templates/role.yaml

generate: setup manifests update-chart ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	cd pkg/sdk && $(CONTROLLER_GEN) object paths="./api/..."

ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
test: manifests generate ## Run tests.
	mkdir -p ${ENVTEST_ASSETS_DIR}
	test -f ${ENVTEST_ASSETS_DIR}/setup-envtest.sh || curl -sSLo ${ENVTEST_ASSETS_DIR}/setup-envtest.sh https://raw.githubusercontent.com/kubernetes-sigs/controller-runtime/v0.8.3/hack/setup-envtest.sh
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); go test ./... -coverprofile cover.out

lint: setup
	go fmt ./...
	go vet ./...
	go mod tidy
	cd ui && yarn run eslint
	helm lint charts/logging-pipeline-plumber

setup: ## Download controller-gen and  locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1) 
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)
	cd ui && yarn install

##@ Build

run: manifests generate ## Run a controller from your host.
	go run ./main.go

build: update-chart
	source scripts/version && scripts/build

##@ Deployment

install: manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

uninstall: manifests ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

deploy: manifests ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	source scripts/version && cd config/manager && $(KUSTOMIZE) edit set image controller=${CONTROLLER_IMAGE_NAME}:${VERSION}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl delete -f -


# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
