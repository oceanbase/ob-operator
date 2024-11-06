##@ Helper

.PHONY: helper-build
helper-build: dashboard-bindata-gen dashboard-doc-gen cli-bindata-gen ## Build oceanbase helper binary.
	@echo "Building helper..."
	CGO_ENABLED=0 GOOS=linux go build -a -o bin/oceanbase-helper ./cmd/oceanbase-helper/main.go