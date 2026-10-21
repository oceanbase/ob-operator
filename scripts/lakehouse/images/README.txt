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

With both base images already present in Docker, run:
  python3 scripts/lakehouse/images/build_images.py \
    --inputs INPUT_DIRECTORY --output NEW_OUTPUT_DIRECTORY

The build uses Docker 26's legacy builder, accepts only Linux/amd64 bases,
does not pull images, and disables networking for builds and smoke tests.
It installs the signed, SHA256-checked tar RPM into Observer without changing
the original command, environment, user or database binary. The obdiag image
uses the official 5.0.1 RPM and the base image's Python 3.6.8 runtime; no pip or
PyYAML installation is needed. RPM package metadata reports "Commercial",
while the obdiag executable reports Mulan PSL v2; both are recorded as supplied
upstream metadata, not a new license grant by this repository.

Outputs
-------
* diagnostics-base-images.tar.gz: Observer, obdiag and oceanbase-helper.
* diagnostics-lock.json: archive SHA256, exact image config IDs, references,
  base config IDs and input provenance.
* Per-image build and smoke logs, including the preserved Observer binary
  SHA256, tar/gzip archive roundtrip and obdiag/helper command checks.

When the Dashboard backend changes, rebuild it from the recorded source revision
after running make dashboard-bindata-gen dashboard-doc-gen, using the repository's
Go version. Record the compiler image ID, build flags and binary SHA256. The QA
Dockerfile.dashboard accepts BASE_IMAGE, SOURCE_REVISION and BACKEND_SHA256;
copy the rebuilt binary into its context as oceanbase-dashboard. The fixed base
must retain a frontend built from the same ui/ source tree (verify this before
reusing it); otherwise rebuild the frontend too.

Release engineering exports the three diagnostic references plus the rebuilt
Dashboard with docker image save into one new archive. Add the Dashboard role,
its image ID, binary SHA256 and source provenance to diagnostics-lock.json, set
binary_rebuild=true and recompute archive_sha256. The package builder validates
that combined archive; it does not construct container images itself.
Docker image IDs can differ on an uncached rebuild because Docker records build
timestamps. Always distribute the config IDs from the actual exported archive;
do not substitute the image tag for an immutable identity.
