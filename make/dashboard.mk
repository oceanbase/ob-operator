##@ Dashboard

PROJECT=oceanbase-dashboard
PROCESSOR=4
PWD ?= $(shell pwd)

DASHBOARD_VERSION ?= 0.3.0
DASHBOARD_IMG ?= quay.io/oceanbase/oceanbase-dashboard:${DASHBOARD_VERSION}
COMMIT_HASH ?= $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP ?= $(shell date '+%Y%m%d%H%M%S')
INJECT_PACKAGE=github.com/oceanbase/ob-operator/internal/dashboard/handler

BUILD_FLAG      := -p $(PROCESSOR) -ldflags="-X '$(INJECT_PACKAGE).Version=$(DASHBOARD_VERSION)' -X '$(INJECT_PACKAGE).CommitHash=$(COMMIT_HASH)' -X '$(INJECT_PACKAGE).BuildTime=$(BUILD_TIMESTAMP)'"
GOBUILD         := GO11MODULE=ON CGO_ENABLED=0 GOOS=linux go build $(BUILD_FLAG)
GOBUILDCOVERAGE := go test -covermode=count -coverpkg="../..." -c .
GOCOVERAGE_FILE := tests/coverage.out
GOCOVERAGE_REPORT := tests/coverage-report
GOTEST          := go test -tags test -covermode=count -coverprofile=$(GOCOVERAGE_FILE) -p $(PROCESSOR)

GOFILES ?= $(shell git ls-files '*.go')
GOTEST_PACKAGES = $(shell go list ./... | grep -v -f tests/excludes.txt)
UNFMT_FILES ?= $(shell gofmt -l -s $(filter-out , $(GOFILES)))

.PHONY: dashboard-doc-gen
dashboard-doc-gen: dashboard-dep-install ## Generate swagger docs
	swag init -g cmd/dashboard/main.go -o internal/dashboard/generated/swagger

.PHONY: dashboard-build
dashboard-build: dashboard-bindata-gen dashboard-doc-gen ## Build oceanbase-dashboard
	$(GOBUILD) -o bin/oceanbase-dashboard ./cmd/dashboard/main.go

.PHONY: dashboard-bindata-gen
dashboard-bindata-gen: ## Generate bindata
	go-bindata -o internal/dashboard/generated/bindata/bindata.go -pkg bindata internal/assets/...

.PHONY: dashboard-clean
dashboard-clean: ## Clean build
	rm -rf bin/oceanbase-dashboard
	go clean -i ./...

.PHONY: dashboard-dep-install
dashboard-dep-install: ## Install dependencies for oceanbase-dashboard
	go install github.com/go-bindata/go-bindata/...@v3.1.2+incompatible
	go install github.com/swaggo/swag/cmd/swag@latest

.PHONY: dashboard-run
dashboard-run: ## Run oceanbase-dashboard in dev mode
	go run $(BUILD_FLAG) ./cmd/dashboard/main.go

.PHONY: dashboard-docker-build
dashboard-docker-build: dashboard-bindata-gen dashboard-doc-gen ## build oceanbase-dashboard image
	docker build -t ${DASHBOARD_IMG} -f build/Dockerfile.dashboard .

.PHONY: dashboard-docker-push
dashboard-docker-push: ## push oceanbase-dashboard image
	docker push ${DASHBOARD_IMG}
