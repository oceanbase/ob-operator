SQL_ANALYZER_VERSION ?= 0.1.0
SQL_ANALYZER_IMG ?= quay.io/oceanbase/sql-analyzer:${SQL_ANALYZER_VERSION}

.PHONY: sql-analyzer-dep-install
sql-analyzer-dep-install: ## Install dependencies for sql-analyzer
	@if [ -z "$(shell command -v swag)" ]; then \
		go install github.com/swaggo/swag/cmd/swag@v1.16.3; \
	fi

.PHONY: sql-analyzer-doc-gen
sql-analyzer-doc-gen: sql-analyzer-dep-install ## Generate swagger docs for sql-analyzer
	swag init -g main.go -o internal/sql-analyzer/generated/swagger -d ./cmd/sql-analyzer,./internal/sql-analyzer

.PHONY: sql-analyzer
sql-analyzer: sql-analyzer-doc-gen
	@echo Building sql-analyzer...
	@mkdir -p bin
ifneq ($(ASAN_ENABLED),)
	@echo "Building with AddressSanitizer (Clang)..."
	CC=clang CGO_ENABLED=1 CGO_LDFLAGS='-fsanitize=address' CGO_CFLAGS='-O0 -g3 -fsanitize=address' go build -asan -o bin/sql-analyzer cmd/sql-analyzer/main.go
else
	@go build -o bin/sql-analyzer cmd/sql-analyzer/main.go
endif

.PHONY: sql-analyzer-image
sql-analyzer-image:
	$(eval DOCKER_BUILD_ARGS :=)
	$(if $(GOPROXY),$(eval DOCKER_BUILD_ARGS := --build-arg GOPROXY=$(GOPROXY)))
	$(if $(ASAN_ENABLED),$(eval DOCKER_BUILD_ARGS := $(DOCKER_BUILD_ARGS) --build-arg ASAN_ENABLED=$(ASAN_ENABLED)))
	docker build $(DOCKER_BUILD_ARGS) -t ${SQL_ANALYZER_IMG} -f build/Dockerfile.sql-analyzer .
