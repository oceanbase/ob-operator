PROJECT=oceanbase-dashboard
PROCESSOR=4
PWD ?= $(shell pwd)

VERSION ?= 1.0.0
COMMIT_HASH ?= $(shell git rev-parse --short HEAD)
BUILD_TIMESTAMP ?= $(shell date '+%Y%m%d%H%M%S')
INJECT_PACKAGE=github.com/oceanbase/oceanbase-dashboard/internal/handler

BUILD_FLAG      := -p $(PROCESSOR) -ldflags="-X '$(INJECT_PACKAGE).Version=$(VERSION)' -X '$(INJECT_PACKAGE).CommitHash=$(COMMIT_HASH)' -X '$(INJECT_PACKAGE).BuildTime=$(BUILD_TIMESTAMP)'"
GOBUILD         := go build $(BUILD_FLAG)
GOBUILDCOVERAGE := go test -covermode=count -coverpkg="../..." -c .
GOCOVERAGE_FILE := tests/coverage.out
GOCOVERAGE_REPORT := tests/coverage-report
GOTEST          := go test -tags test -covermode=count -coverprofile=$(GOCOVERAGE_FILE) -p $(PROCESSOR)


GOFILES ?= $(shell git ls-files '*.go')
GOTEST_PACKAGES = $(shell go list ./... | grep -v -f tests/excludes.txt)
UNFMT_FILES ?= $(shell gofmt -l -s $(filter-out , $(GOFILES)))
