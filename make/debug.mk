# All targets here are phony
##@ Debug

.PHONY: connect root-pwd run-delve install-delve

NS ?= default
LOG_LEVEL ?= 2

connect: root-pwd
	$(eval nodeHost = $(shell kubectl -n ${NS} get pods -l ref-obcluster=$(clusterName) -o jsonpath='{.items[0].status.podIP}'))
ifdef TENANT
	$(eval secretName = $(shell kubectl -n ${NS} get obtenant ${TENANT} -o jsonpath='{.status.credentials.root}'))
	$(eval tenantName = $(shell kubectl -n ${NS} get obtenant ${TENANT} -o jsonpath='{.spec.tenantName}'))
	$(if $(strip $(secretName)), $(eval pwd = $(shell kubectl -n ${NS} get secret $(secretName) -o jsonpath='{.data.password}' | base64 -d)), )
	$(if $(strip $(pwd)), mysql -h$(nodeHost) -P2881 -A -uroot@$(tenantName) -p$(pwd) -Doceanbase, mysql -h$(nodeHost) -P2881 -A -uroot@$(tenantName) -Doceanbase)
else
	mysql -h$(nodeHost) -P2881 -A -uroot -p -Doceanbase -p$(pwd)
endif

root-pwd:
ifdef CLUSTER
	$(eval clusterName = ${CLUSTER})
else
	$(eval clusterName = $(shell kubectl -n ${NS} get obcluster -o jsonpath='{.items[0].metadata.name}'))
endif
	@echo clusterName $(clusterName)
	$(eval secretName = $(shell kubectl -n ${NS} get obcluster $(clusterName) -o jsonpath='{.spec.userSecrets.root}'))
	$(eval nodeHost = $(shell kubectl -n ${NS} get pods -l ref-obcluster=$(clusterName) -o jsonpath='{.items[0].status.podIP}'))
	$(if $(strip $(secretName)), $(eval pwd = $(shell kubectl -n ${NS} get secret $(secretName) -o jsonpath='{.data.password}' | base64 -d)), )
	@echo root pwd of sys of cluster '$(clusterName)' is $(pwd)

## Delve is a debugger for the Go programming language. More info: https://github.com/go-delve/delve
run-delve: generate fmt vet manifests ## Run with Delve for development purposes against the configured Kubernetes cluster in ~/.kube/config 
	go build -gcflags "all=-trimpath=$(shell go env GOPATH)" -o bin/manager cmd/main.go
	DISABLE_WEBHOOKS=true DISABLE_TELEMETRY=true dlv --listen=:2345 --headless=true --api-version=2 --accept-multiclient exec ./bin/manager --continue -- -log-verbosity=${LOG_LEVEL}

install-delve: ## Install delve, a debugger for the Go programming language. More info: https://github.com/go-delve/delve
	go install github.com/go-delve/delve/cmd/dlv@master

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	@mkdir -p testreports/covdata
	CGO_ENABLED=1 GOCOVERDIR=testreports/covdata DISABLE_WEBHOOKS=true DISABLE_TELEMETRY=true go run -cover -covermode=atomic ./cmd/main.go --log-verbosity=${LOG_LEVEL} 
