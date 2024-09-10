##@ cli

PROJECT=oceanbase-cli
PROCESSOR=4
PWD ?= $(shell pwd)
CLI_VERSION ?= 0.1.0
CLI_IMG ?= quay.io/oceanbase/oceanbase-cli:${CLI_VERSION}

BUILD_FLAG      := -p $(PROCESSOR)
GOBUILD := GO11MODULE=ON CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAG)
.PHONY: cli-build
cli-build: cli-bindata-gen cli-dep-install # Build oceanbase-cli
	$(GOBUILD) -o bin/obocli cmd/cli/main.go

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
	go run $(BUILD_FLAG) ./cmd/cli/main.go

.PHONY: cli-docker-build
cli-docker-build: cli-bindata-gen ## build oceanbase-cli image
	docker build -t ${CLI_IMG} -f build/Dockerfile.cli .

.PHONY: cli-docker-push
cli-docker-push: ## push oceanbase-cli image
	docker push ${CLI_IMG}
