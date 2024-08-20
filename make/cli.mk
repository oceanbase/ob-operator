##@ cli

PROJECT=oceanbase-cli
PROCESSOR=4
PWD ?= $(shell pwd)

.PHONY: cli-build
cli-build: cli-dep-install # Build oceanbase-cli
	go build -o bin/obocli cmd/cli/main.go

.PHONY: cli-clean
cli-clean: # Clean build
	rm -rf bin/obocli
	go clean -i ./...

.PHONY : cli-dep-install
cli-dep-install: # Install oceanbase-cli deps
	go install github.com/spf13/cobra

.PHONY : cli-run
cli-run: ## Run oceanbase-cli in dev mode
	go run ./cmd/cli/main.go

