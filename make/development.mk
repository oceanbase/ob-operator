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
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test -timeout 60m  -v ./... -coverprofile cover.out

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