##@ Development

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
	--output-dir=testreports --keep-going --json-report=report.json --label-filter='!long-run'
	
.PHONY: test-all
test-all: manifests generate fmt vet envtest ## Run tests to get cpu and mem profile
	TEST_USE_EXISTING_CLUSTER=true TELEMETRY_REPORT_HOST=http://openwebapi.test.alipay.net \
	DISABLE_TELEMETRY=true LOG_VERBOSITY=0 \
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" \
	go run github.com/onsi/ginkgo/v2/ginkgo -r --coverprofile=cover.profile --cpuprofile=cpu.profile --memprofile=mem.profile --cover \
	--output-dir=testreports --keep-going --json-report=report.json --label-filter='$(CASE_LABEL_FILTERS)'

REPORT_PORT ?= 8480

.PHONY: coverage
coverage: test ## Generate test reports
	@echo "generating test reports..."
	@go tool cover -html=testreports/cover.profile -o testreports/cover.html
	@cd testreports && python3 -m http.server --bind 0.0.0.0 $(REPORT_PORT)

.PHONY: run-coverage
run-coverage:  ## Generate integration test coverage report.
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