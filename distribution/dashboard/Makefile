include make/*

.PHONY: all test clean build 

default: clean fmt generate-bindata doc build

doc: dep-install
	swag init

build:
	$(GOBUILD) -o bin/oceanbase-dashboard ./main.go

test:
	$(GOTEST) $(GOTEST_PACKAGES)

fmt:
	@gofmt -s -w $(filter-out , $(GOFILES))

fmt-check:
	@if [ -z "$(UNFMT_FILES)" ]; then \
		echo "gofmt check passed"; \
		exit 0; \
    else \
    	echo "gofmt check failed, not formatted files:"; \
    	echo "$(UNFMT_FILES)" | tr -s " " "\n"; \
    	exit 1; \
    fi

generate-bindata:
	go-bindata -o bindata/bindata.go -pkg bindata assets/...

tidy:
	go mod tidy

vet:
	go vet $$(go list ./...)

clean:
	rm -rf bin/oceanbase-dashboard
	go clean -i ./...

dep-install:
	go install github.com/go-bindata/go-bindata/...@v3.1.2+incompatible
	go install github.com/swaggo/swag/cmd/swag@latest

dev:
	go run $(BUILD_FLAG) main.go

run: build
	./bin/oceanbase-dashboard
