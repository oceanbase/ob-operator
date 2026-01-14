##@ okctl

BINARY_NAME ?= okctl
CLI_VERSION ?= 0.1.0
GOARCH ?=$(shell uname -m)
GOOS ?= $(shell uname -s | tr LD ld)
VERSION_INJECT_PACKAGE=github.com/oceanbase/ob-operator/internal/cli/cmd/version
LOGGER_HEAD_INJECT_PACKAGE=github.com/oceanbase/ob-operator/internal/cli/utils
BINARY_NAME_INJECT_PACKAGE=github.com/oceanbase/ob-operator/internal/cli
COMMIT_HASH ?= $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP ?= $(shell date '+%Y%m%d%H%M%S')

# If GOARCH not specified, set GOARCH based on the detected architecture
ifeq ($(GOARCH),x86_64)
    GOARCH = amd64
else ifeq ($(GOARCH),aarch64)
    GOARCH = arm64
endif

# build flags
CLI_BUILD_FLAGS = -p $(PROCESSOR) -ldflags="-X '$(VERSION_INJECT_PACKAGE).Version=$(CLI_VERSION)' -X '$(VERSION_INJECT_PACKAGE).CommitHash=$(COMMIT_HASH)' -X '$(VERSION_INJECT_PACKAGE).BuildTime=$(BUILD_TIMESTAMP)' -X '$(BINARY_NAME_INJECT_PACKAGE).BinaryName=$(BINARY_NAME)'  -X '$(LOGGER_HEAD_INJECT_PACKAGE).BinaryName=$(BINARY_NAME)'"

# build args
CLI_BUILD := GO11MODULE=ON CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(CLI_BUILD_FLAGS)

# build dir for cli
BUILD_DIR?=bin/

.PHONY: okctl
okctl:  ## Build oceanbase-cli
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GOARCH)..."
	$(CLI_BUILD) -o $(BUILD_DIR)$(BINARY_NAME) cmd/cli/main.go

.PHONY: okctl-clean
okctl-clean: ## Clean build
	rm -rf $(RELEASE_DIR)/$(BINARY_NAME)
	go clean -i ./...

.PHONY : okctl-dep-install
okctl-dep-install: ## Install oceanbase-cli deps
	@if [ -z "$(shell command -v go-bindata)" ]; then \
		go install github.com/go-bindata/go-bindata/...@v3.1.2+incompatible; \
	fi
	@if [ -z "$(shell command -v cobra)" ]; then \
		go install github.com/spf13/cobra; \
	fi
	
.PHONY : okctl-run
okctl-run: ## Run oceanbase-cli in dev mode
	go run -p 4 ./cmd/cli/main.go