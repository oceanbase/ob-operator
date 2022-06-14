# Ensure Make is run with bash shell as some syntax below is bash-specific
SHELL:=/usr/bin/env bash

.DEFAULT_GOAL := help

# Code
VERSION := operator-v3.1.3-10100042022051821

# Define Docker related variables. Releases should modify and double check these vars.
REGISTRY := oceanbase
IMAGE := oceanbase-ce

# Use GOPROXY environment variable if set
GOPROXY := $(shell go env GOPROXY)
ifeq ($(GOPROXY),)
GOPROXY := https://goproxy.cn
endif

export GOPROXY
# Active module mode, as we use go modules to manage dependencies
export GO111MODULE=on

run:
	go run ../../cmd/cable/app.go

server:
	@echo "version: $(VERSION)"
	cd ../../ && docker build -t $(REGISTRY)/$(IMAGE):$(VERSION) -f scripts/observer/Dockerfile .
	docker push $(REGISTRY)/$(IMAGE):$(VERSION)
