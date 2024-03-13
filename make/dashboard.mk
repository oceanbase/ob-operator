##@ Dashboard

PROJECT=oceanbase-dashboard
PROCESSOR=4
PWD ?= $(shell pwd)

DASHBOARD_VERSION ?= 1.0.0
COMMIT_HASH ?= $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP ?= $(shell date '+%Y%m%d%H%M%S')
INJECT_PACKAGE=github.com/oceanbase/oceanbase-operator/internal/handler

BUILD_FLAG      := -p $(PROCESSOR) -ldflags="-X '$(INJECT_PACKAGE).Version=$(DASHBOARD_VERSION)' -X '$(INJECT_PACKAGE).CommitHash=$(COMMIT_HASH)' -X '$(INJECT_PACKAGE).BuildTime=$(BUILD_TIMESTAMP)'"
GOBUILD         := go build $(BUILD_FLAG)
GOBUILDCOVERAGE := go test -covermode=count -coverpkg="../..." -c .
GOCOVERAGE_FILE := tests/coverage.out
GOCOVERAGE_REPORT := tests/coverage-report
GOTEST          := go test -tags test -covermode=count -coverprofile=$(GOCOVERAGE_FILE) -p $(PROCESSOR)

GOFILES ?= $(shell git ls-files '*.go')
GOTEST_PACKAGES = $(shell go list ./... | grep -v -f tests/excludes.txt)
UNFMT_FILES ?= $(shell gofmt -l -s $(filter-out , $(GOFILES)))

.PHONY: dashboard-doc
dashboard-doc: dep-install ## Generate swagger docs
	swag init -g cmd/dashboard/main.go -o internal/dashboard/generated/swagger

.PHONY: build-dashboard
build-dashboard: ## Build oceanbase-dashboard
	$(GOBUILD) -o bin/oceanbase-dashboard ./cmd/dashboard/main.go

.PHONY: gen-bindata
gen-bindata: ## Generate bindata
	go-bindata -o internal/dashboard/generated/bindata/bindata.go -pkg bindata internal/assets/...

.PHONY: clean
clean: ## Clean build
	rm -rf bin/oceanbase-dashboard
	go clean -i ./...

.PHONY: dep-install
dep-install: ## Install dependencies for oceanbase-dashboard
	go install github.com/go-bindata/go-bindata/...@v3.1.2+incompatible
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: dev-dashboard
dev-dashboard: ## Run oceanbase-dashboard in dev mode
	go run $(BUILD_FLAG) ./cmd/dashboard/main.go