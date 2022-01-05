#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# get code-generator
K8S_VERSION=v0.22.1
go get k8s.io/code-generator@$K8S_VERSION
go get k8s.io/client-go@$K8S_VERSION
go get k8s.io/apimachinery@$K8S_VERSION
go get sigs.k8s.io/controller-runtime@v0.10.0
go mod vendor
chmod +x vendor/k8s.io/code-generator/generate-groups.sh

# corresponding to go mod init <module>
MODULE=github.com/oceanbase/ob-operator
# client package
OUTPUT_PKG=pkg/kubeclient
# api package
APIS_PKG=apis
# group-version
GROUP_VERSION=cloud:v1

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ./vendor/k8s.io/code-generator 2>/dev/null || echo ../code-generator)}

rm -rf ${OUTPUT_PKG}/clientset
rm -rf ${OUTPUT_PKG}/listers
rm -rf ${OUTPUT_PKG}/informers

bash "${CODEGEN_PKG}"/generate-groups.sh "client,lister,informer" \
  ${MODULE}/${OUTPUT_PKG} \
  ${MODULE}/${APIS_PKG} \
  ${GROUP_VERSION} \
  --go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt

rm -rf vendor/
