##@ Development

LOG_LEVEL ?= 2

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: controller-gen ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

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
	go run github.com/onsi/ginkgo/v2/ginkgo -r --coverprofile=cover.profile --cpuprofile=cpu.profile --memprofile=mem.profile --cover \
	--output-dir=testreports --keep-going --json-report=report.json --label-filter='!long-run' --skip-package './distribution'
	
.PHONY: test-all
test-all: manifests generate fmt vet envtest ## Run all tests including long-run ones.
	TEST_USE_EXISTING_CLUSTER=true TELEMETRY_REPORT_HOST=http://openwebapi.test.alipay.net \
	DISABLE_TELEMETRY=true LOG_VERBOSITY=0 \
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
	go run github.com/onsi/ginkgo/v2/ginkgo -r --coverprofile=cover.profile --cpuprofile=cpu.profile --memprofile=mem.profile --cover \
	--output-dir=testreports --keep-going --json-report=report.json --label-filter='$(CASE_LABEL_FILTERS)' --skip-package './distribution'

REPORT_PORT ?= 8480

.PHONY: unit-coverage
unit-coverage: test ## Generate unit test coverage report.
	@echo "generating test reports..."
	@go tool cover -html=testreports/cover.profile -o testreports/unit.html
	@cd testreports && python3 -m http.server --bind 0.0.0.0 $(REPORT_PORT)

.PHONY: intg-coverage
intg-coverage:  ## Generate integration test coverage report.
	@go tool covdata textfmt -i=testreports/covdata -o=testreports/covdata.txt
	@go tool cover -html=testreports/covdata.txt -o testreports/integration.html
	@go tool cover -html=testreports/cover.profile -o testreports/unit.html
	@cd testreports && python3 -m http.server --bind 0.0.0.0 $(REPORT_PORT)

.PHONY: GOLANGCI_LINT
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
$(GOLANGCI_LINT):
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@${GOLANG_CI_VERSION}

.PHONY: lint
lint: $(GOLANGCI_LINT) ## Run linting.
	$(GOLANGCI_LINT) run -v --timeout=10m

.PHONY: commit-hook
commit-hook: $(GOLANGCI_LINT) ## Install commit hook.
	touch .git/hooks/pre-commit
	chmod +x .git/hooks/pre-commit
	echo "#!/bin/sh" > .git/hooks/pre-commit
	echo "make lint" >> .git/hooks/pre-commit

.PHONY: run-delve
run-delve: generate fmt vet manifests ## Run with Delve for development purposes against the configured Kubernetes cluster in ~/.kube/config 
	go build -gcflags "all=-trimpath=$(shell go env GOPATH)" -o bin/manager cmd/main.go
	DISABLE_WEBHOOKS=true DISABLE_TELEMETRY=true dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/manager --continue -- -log-verbosity=${LOG_LEVEL}

.PHONY: run-local
run-local: manifests generate fmt vet ## Run a controller on your local host, with configurations in ~/.kube/config
	@mkdir -p testreports/covdata
	CGO_ENABLED=1 GOCOVERDIR=testreports/covdata DISABLE_WEBHOOKS=true DISABLE_TELEMETRY=true go run -cover -covermode=atomic ./cmd/main.go --log-verbosity=${LOG_LEVEL} 
