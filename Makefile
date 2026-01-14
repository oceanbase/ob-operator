# Get the number of processors
ifeq ($(shell uname -s), Linux)
	PROCESSOR = $(shell nproc)
else ifeq ($(shell uname -s), Darwin)
	PROCESSOR = $(shell sysctl -n hw.ncpu)
else
	PROCESSOR = 4
endif

include make/*

VERSION ?= 2.3.4
# Image URL to use all building/pushing image targets
IMG ?= quay.io/oceanbase/ob-operator:${VERSION}
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.26.1

YQ_DOWNLOAD_URL = https://github.com/mikefarah/yq/releases/download/v4.35.1/yq_linux_amd64
SEMVER_DOWNLOAD_URL = https://raw.githubusercontent.com/fsaintjacques/semver-tool/master/src/semver

GOLANG_CI_VERSION ?= v1.54.2

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif
GOPROXY ?= https://goproxy.io,direct
GOSUMDB ?= sum.golang.org
RACE ?= ''

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: operator

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)
