##@ Development

LOG_LEVEL ?= 2

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen init-generator ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations; Generate task registrations for resource manager.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."
	go generate ./...

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	TEST_USE_EXISTING_CLUSTER=true TELEMETRY_REPORT_HOST=http://openwebapi.test.alipay.net \
	DISABLE_TELEMETRY=true LOG_VERBOSITY=0 \
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
	go run github.com/onsi/ginkgo/v2/ginkgo -r --covermode=atomic --coverprofile=cover.profile --cpuprofile=cpu.profile --memprofile=mem.profile --cover \
	--output-dir=testreports --keep-going --json-report=report.json --label-filter='!long-run' --skip-package './distribution'
	
.PHONY: test-all
test-all: manifests generate fmt vet envtest ## Run all tests including long-run ones.
	TEST_USE_EXISTING_CLUSTER=true TELEMETRY_REPORT_HOST=http://openwebapi.test.alipay.net \
	DISABLE_TELEMETRY=true LOG_VERBOSITY=0 \
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
	go run github.com/onsi/ginkgo/v2/ginkgo -r --covermode=atomic --coverprofile=cover.profile --cpuprofile=cpu.profile --memprofile=mem.profile --cover \
	--output-dir=testreports --keep-going --json-report=report.json --label-filter='$(CASE_LABEL_FILTERS)' --skip-package './distribution'

REPORT_PORT ?= 8480

.PHONY: unit-coverage
unit-coverage: test ## Generate unit test coverage report.
	@echo "generating test reports..."
	@go tool cover -html=testreports/cover.profile -o testreports/unit.html
	@cd testreports && python3 -m http.server --bind 0.0.0.0 $(REPORT_PORT)

.PHONY: intg-coverage
intg-coverage:  ## Generate integration test coverage report.
	@go tool cover -html=testreports/cover.profile -o testreports/unit.html
	@go tool covdata textfmt -i=testreports/covdata -o=testreports/covdata.txt
	@go tool cover -html=testreports/covdata.txt -o testreports/integration.html
	@cd testreports && python3 -m http.server --bind 0.0.0.0 $(REPORT_PORT)

.PHONY: GOLANGCI_LINT
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
$(GOLANGCI_LINT):
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANG_CI_VERSION}

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Run linting.
	$(GOLANGCI_LINT) run -v --timeout=10m --max-same-issues=1000

.PHONY: ADD_LICENSE_CHECKER
ADD_LICENSE_CHECKER ?= $(LOCALBIN)/addlicense
$(ADD_LICENSE_CHECKER):
	GOBIN=$(LOCALBIN) go install github.com/google/addlicense@latest

.PHONY: license-check
license-check: $(ADD_LICENSE_CHECKER) ## Check whether all license headers are present.
	find . -type f -name "*.go" -not -path "./distribution/*" -not -path "**/generated/*" | xargs $(ADD_LICENSE_CHECKER) -check

.PHONY: commit-hook
commit-hook: $(GOLANGCI_LINT) ## Install commit hook.
	touch .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	echo "#!/bin/sh" > .git/hooks/pre-commit
	echo "make lint" >> .git/hooks/pre-commit
	echo "make export-operator export-charts" >> .git/hooks/pre-commit

.PHONY: run-delve
run-delve: fmt vet manifests ## Run with Delve for development purposes against the configured Kubernetes cluster in ~/.kube/config 
	go build -gcflags "all=-trimpath=$(shell go env GOPATH)" -o bin/manager cmd/operator/main.go
	DISABLE_WEBHOOKS=true dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/manager --continue -- --log-verbosity=${LOG_LEVEL}

.PHONY: run-local
run-local: manifests fmt vet ## Run a controller on your local host, with configurations in ~/.kube/config
	@mkdir -p testreports/covdata
	CGO_ENABLED=0 GOCOVERDIR=testreports/covdata DISABLE_WEBHOOKS=true go run -cover -covermode=atomic ./cmd/operator/main.go --log-verbosity=${LOG_LEVEL} 
