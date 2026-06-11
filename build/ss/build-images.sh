#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RPM_DIR="${RPM_DIR:?Set RPM_DIR to directory containing SS RPMs}"
TAG_LS="${TAG_LS:-1.2.0}"
TAG_OB="${TAG_OB:-4.6.0.0-ss}"

info() { echo ">>> $*"; }

DOCKER="${DOCKER:-docker}"

[[ -d "${RPM_DIR}" ]] || { echo "RPM dir not found: ${RPM_DIR}" >&2; exit 1; }

CTX=$(mktemp -d)
trap 'rm -rf "${CTX}"' EXIT

info "Prepare build context from ${RPM_DIR}"
cp "${RPM_DIR}"/*.rpm "${CTX}/"
cp "${SCRIPT_DIR}/../scripts/observer-prestart-wrapper.sh" "${CTX}/"

HELPER_SRC="${OB_OPERATOR_ROOT:-$(git -C "${SCRIPT_DIR}" rev-parse --show-toplevel 2>/dev/null || echo "${SCRIPT_DIR}/../..")}"
GOROOT_TOOLCHAIN="${GOROOT_TOOLCHAIN:-$(go env GOROOT)}"
info "Building oceanbase-helper from ${HELPER_SRC} ..."
(
  cd "${HELPER_SRC}"
  export GOROOT="${GOROOT_TOOLCHAIN}"
  export PATH="${GOROOT}/bin:${PATH}"
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o "${CTX}/oceanbase-helper" ./cmd/oceanbase-helper/
)

info "Building oblogservice:${TAG_LS} ..."
cp "${SCRIPT_DIR}/Dockerfile.oblogservice" "${CTX}/Dockerfile"
${DOCKER} build -t "oblogservice:${TAG_LS}" "${CTX}"

info "Building oceanbase-ss:${TAG_OB} ..."
cp "${SCRIPT_DIR}/Dockerfile.observer-ss" "${CTX}/Dockerfile"
${DOCKER} build -t "oceanbase-ss:${TAG_OB}" "${CTX}"

info "Done."
${DOCKER} images | grep -E "oblogservice|oceanbase-ss"
