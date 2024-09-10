##@ cli

BINARY_NAME ?= obocli
CLI_VERSION ?= 0.1.0
GOARCH ?=$(shell uname -m)
GOOS ?= $(shell uname -s | tr LD ld)

# If GOARCH not specified, set GOARCH based on the detected architecture
ifeq ($(GOARCH),x86_64)
    GOARCH = amd64
else ifeq ($(GOARCH),aarch64)
    GOARCH = arm64
endif

# build args
CLI_BUILD := GO11MODULE=ON CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build

# build dir for obocli 
BUILD_DIR?=bin/

BUILD_FLAG      := -p $(PROCESSOR)
GOBUILD := GO11MODULE=ON CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAG)
.PHONY: cli-build
cli-build: cli-dep-install cli-bindata-gen # Build oceanbase-cli
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	$(CLI_BUILD) -o $(BUILD_DIR)$(BINARY_NAME) cmd/cli/main.go

.PHONY: cli-bindata-gen
cli-bindata-gen: # Generate bindata
	go-bindata -o internal/cli/generated/bindata/bindata.go -pkg bindata internal/assets/cli-templates/...

.PHONY: cli-clean
cli-clean: # Clean build
	rm -rf $(RELEASE_DIR)/$(BINARY_NAME)
	go clean -i ./...

.PHONY : cli-dep-install
cli-dep-install: # Install oceanbase-cli deps
	go install github.com/spf13/cobra
	go install github.com/go-bindata/go-bindata/...@v3.1.2+incompatible
	
.PHONY : cli-run
cli-run: ## Run oceanbase-cli in dev mode
	go run -p 4 ./cmd/cli/main.go
