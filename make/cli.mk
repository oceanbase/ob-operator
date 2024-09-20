##@ cli

PROJECT=oceanbase-cli
CLI_VERSION ?= 0.1.0
CLI_IMG ?= quay.io/oceanbase/oceanbase-cli:${CLI_VERSION}

CLI_BUILD := GO11MODULE=ON CGO_ENABLED=0 GOOS=linux go build
.PHONY: cli-build
cli-build: cli-bindata-gen cli-dep-install # Build oceanbase-cli
	$(CLI_BUILD) -o bin/obocli cmd/cli/main.go

.PHONY: cli-bindata-gen
cli-bindata-gen: # Generate bindata
	go-bindata -o internal/cli/generated/bindata/bindata.go -pkg bindata internal/assets/cli-templates/...

.PHONY: cli-clean
cli-clean: # Clean build
	rm -rf bin/obocli
	go clean -i ./...

.PHONY : cli-dep-install
cli-dep-install: # Install oceanbase-cli deps
	go install github.com/spf13/cobra
	go install github.com/go-bindata/go-bindata/...@v3.1.2+incompatible
	
.PHONY : cli-run
cli-run: ## Run oceanbase-cli in dev mode
	go run -p 4 ./cmd/cli/main.go
