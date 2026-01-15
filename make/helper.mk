##@ Helper

.PHONY: helper
helper: ## Build oceanbase helper binary.
	@echo "Building helper..."
	CGO_ENABLED=0 GOOS=linux go build -p $(PROCESSOR) -a -o bin/oceanbase-helper ./cmd/oceanbase-helper/main.go
