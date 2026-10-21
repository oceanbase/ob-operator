Diagnostic image build (Linux/amd64)

These are release-engineering recipes, not additional end-user installation
steps. The generated lakehouse package imports these images automatically.

Inputs and provenance
---------------------
inputs.lock.json fixes the base image config IDs, upstream runtime digest,
official RPM URLs and SHA256 values, signing key, helper source revision and
helper binary SHA256. Prepare these exact inputs in a separate directory.
The Observer base comes from the existing QA image archive; its database
binary is not rebuilt or replaced. Keep that original archive as a build input.

Build the helper from a clean git archive of helper_revision using Go 1.25.0:
  CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false \
    -ldflags=-buildid= -o INPUT_DIRECTORY/oceanbase-helper \
    ./cmd/oceanbase-helper/main.go
Run this command from the extracted source root and use an absolute output
directory. Verify the resulting SHA256 against inputs.lock.json.

With the pinned runtime base already present in Docker, run:
  python3 scripts/lakehouse/images/build_images.py \
    --inputs INPUT_DIRECTORY --output NEW_OUTPUT_DIRECTORY \
    --source-revision REVIEWED_TRANSPORT_COMMIT

The build uses Docker 26's legacy builder, accepts only Linux/amd64 bases,
does not pull images, and disables networking for builds and smoke tests.
The optional --include-observer path also needs the pinned Observer base. It
installs the signed, SHA256-checked tar RPM without changing the original
command, environment, user or database binary. The obdiag image
uses the official 5.0.1 RPM and the base image's Python 3.6.8 runtime; no pip or
PyYAML installation is needed. A pinned-source compatibility patch changes only
Kubernetes log archive creation/download: bounded exec requires a successful
exit status, base64 chunks carry a checked size/SHA256, and publication is atomic.
Existing tar/gzip create archives when available; otherwise the target's Python
standard library creates them. Download always requires target Python >=3.6
(platform-python or python3). Its probe creates and reads a gzip tar archive
entirely in memory. No packages, images, executables or database configuration
are changed on the target. Other Kubernetes downloads and SSH/Docker paths
retain upstream behavior.

RPM package metadata reports "Commercial",
while the obdiag executable reports Mulan PSL v2; both are recorded as supplied
upstream metadata, not a new license grant by this repository.

Outputs
-------
* diagnostics-base-images.tar.gz: obdiag and oceanbase-helper. The optional
  --include-observer flag includes the tar-enabled Observer for new installs.
  Existing shared-storage Observer images must stay unchanged.
* diagnostics-lock.json: archive SHA256, exact image config IDs, references,
  base config IDs and input provenance.
* Per-image build and smoke logs, including the preserved Observer binary
  SHA256/tar-gzip roundtrip when requested, and obdiag/helper command checks.

When rebuilding only obdiag, pass --reuse-helper-image-id with the previously
verified helper config ID. The builder checks that exact identity, its base
layers, CLI and binary SHA256 without rebuilding or retagging it. Reuse the
previously verified Dashboard image too when its backend/frontend are unchanged.

When the Dashboard backend changes, rebuild it from the recorded source revision
after running make dashboard-bindata-gen dashboard-doc-gen, using the repository's
Go version. Record the compiler image ID, build flags and binary SHA256. The QA
Dockerfile.dashboard accepts BASE_IMAGE, SOURCE_REVISION and BACKEND_SHA256;
copy the rebuilt binary into its context as oceanbase-dashboard. The fixed base
must retain a frontend built from the same ui/ source tree (verify this before
reusing it); otherwise rebuild the frontend too.

Release engineering exports the selected diagnostic references plus the rebuilt
Dashboard with docker image save into one new archive. Add the Dashboard role,
its image ID, binary SHA256 and source provenance to diagnostics-lock.json, set
binary_rebuild=true and recompute archive_sha256. The package builder validates
that combined archive; it does not construct container images itself.
Docker image IDs can differ on an uncached rebuild because Docker records build
timestamps. Always distribute the config IDs from the actual exported archive;
do not substitute the image tag for an immutable identity.
