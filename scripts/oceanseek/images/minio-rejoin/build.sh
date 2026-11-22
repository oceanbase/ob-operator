#!/usr/bin/env bash
set -euo pipefail

if [[ $# != 1 ]]; then
  echo "Usage: $0 IMAGE_TAG" >&2
  exit 2
fi

tag=$1
: "${BASE_IMAGE:?Set BASE_IMAGE to the immutable digest or local image ID of your verified MinIO runtime}"
[[ $BASE_IMAGE =~ (^|@)sha256:[a-f0-9]{64}$ ]] || { echo "BASE_IMAGE must use an immutable sha256 digest" >&2; exit 2; }
recipe="$(cd "$(dirname "$0")" && pwd)"
tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT
cd "$tmp"

export GOTOOLCHAIN=go1.24.6 GOWORK=off
version=v0.0.0-20250907161309-07c3a429bfed
commit=07c3a429bfed433e49018cb0f78a52145d4bedeb
release=2025-09-07T16-13-09Z
sum=h1:deu0m9BiyqnMbGWfJi1PpVyotSXlDnTmRIme1cUXnxo=
go mod download -json "github.com/minio/minio@$version" > module.json
source_dir="$(python3 -c 'import json, sys; m = json.load(sys.stdin); assert m["Sum"] == sys.argv[1], "unexpected MinIO module checksum"; print(m["Dir"])' "$sum" < module.json)"
cp -R "$source_dir" source
chmod -R u+w source
cd source
patch -p1 < "$recipe/rejoin.patch"
go test ./cmd -run '^TestOceanSeekRejoin' -count=1

mkdir "$tmp/image"
flags="-s -w -X github.com/minio/minio/cmd.Version=$release"
flags+=" -X github.com/minio/minio/cmd.ReleaseTag=RELEASE.$release"
flags+=" -X github.com/minio/minio/cmd.CommitID=$commit"
flags+=" -X github.com/minio/minio/cmd.ShortCommitID=${commit:0:12}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false -ldflags "$flags" -o "$tmp/image/minio" .
docker build --build-arg "BASE_IMAGE=$BASE_IMAGE" --platform linux/amd64 -f "$recipe/Dockerfile" -t "$tag" "$tmp/image"
