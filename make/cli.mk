##@ cli

BINARY_NAME ?= obocli
CLI_VERSION ?= 0.1.0
GOARCH ?=$(shell uname -m)
GOOS ?= $(shell uname -s | tr LD ld)
VERSION_INJECT_PACKAGE=github.com/oceanbase/ob-operator/internal/cli/cmd/version
COMMIT_HASH ?= $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP ?= $(shell date '+%Y%m%d%H%M%S')

# If GOARCH not specified, set GOARCH based on the detected architecture
ifeq ($(GOARCH),x86_64)
    GOARCH = amd64
else ifeq ($(GOARCH),aarch64)
    GOARCH = arm64
endif

CLI_BUILD_FLAGS = -p 4 -ldflags="-X '$(VERSION_INJECT_PACKAGE).Version=$(CLI_VERSION)' -X '$(VERSION_INJECT_PACKAGE).CommitHash=$(COMMIT_HASH)' -X '$(VERSION_INJECT_PACKAGE).BuildTime=$(BUILD_TIMESTAMP)' "
# build args
CLI_BUILD := GO11MODULE=ON CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(CLI_BUILD_FLAGS)

# build dir for obocli 
BUILD_DIR?=bin/

.PHONY: cli-build
cli-build: cli-bindata-gen dashboard-doc-gen dashboard-bindata-gen ## Build oceanbase-cli
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	$(CLI_BUILD) -o $(BUILD_DIR)$(BINARY_NAME) cmd/cli/main.go

.PHONY: cli-bindata-gen
cli-bindata-gen: cli-dep-install ## Generate bindata
	go-bindata -o internal/cli/generated/bindata/bindata.go -pkg bindata internal/assets/cli-templates/...

.PHONY: cli-clean
cli-clean: ## Clean build
	rm -rf $(RELEASE_DIR)/$(BINARY_NAME)
	go clean -i ./...

.PHONY : cli-dep-install
cli-dep-install: ## Install oceanbase-cli deps
	@if [ -z "$(shell command -v go-bindata)" ]; then \
		go install github.com/go-bindata/go-bindata/...@v3.1.2+incompatible; \
	fi
	@if [ -z "$(shell command -v cobra)" ]; then \
		go install github.com/spf13/cobra; \
	fi
	
.PHONY : cli-run
cli-run: ## Run oceanbase-cli in dev mode
	go run -p 4 ./cmd/cli/main.go
