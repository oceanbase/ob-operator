# Native OceanSeek release packaging

This workflow prepares the **same release** for Docker Compose and native Helm.
It does not deploy Kubernetes, start a database, pull images, or publish images.
The earlier `build-package.py` and `build-online-package.py` workflows reproduce
historical installer bundles; they are not the native release entry point.

## OBProxy-enabled development charts

The current source chart installs OBProxy by default. Use the current
`release-images.lock.json`, which includes the Helm-only `obproxy` image role.

## Offline tags and registry publication

Load the delivery archive into Docker or into containerd on every eligible
Kubernetes node. Tags under `localhost/oceanseek` identify locally imported
images. To deploy via a registry, publish the images to a repository you
control and pass its exact references, config IDs and manifest digests in
`--lock`. For online bundles, supply per-role push/pull evidence using
`--publication-report`.

Pass the reviewed release lock explicitly with `--lock`. For public delivery,
pass `--public-release` to the native or Helm builder to reject non-public registry
endpoints. Local acceptance and air-gapped builds may use their own registry.
The packaged lock contains deployment identity and retains the original image
config and manifest digests. Build records contain the checksum of the input
lock, which can include private cache metadata and stays with the publisher.
New builds install
`/usr/local/bin/oceanseek-bootstrap` and `/oceanseek-readiness`. Verify both
paths in the chosen platform images before recording their config IDs. A separate
readiness image can be supplied as the Helm-only `readiness` role.
Every component has one role, canonical image reference (including a fixed tag),
full Docker config image ID, and platform. Both delivery formats use that record.
A role cannot select a different Observer build simply because it runs in Helm.

The release contains 15 roles:

| Delivery | Roles |
| --- | --- |
| Shared by Compose and Helm | `obagent`, `observer`, `logservice`, `minio`, `mc`, `prometheus`, `alertmanager` |
| Compose only | `bootstrap` |
| Helm only | `operator`, `dashboard`, `obproxy`, `obdiag`, `oceanbase-helper`, `utility`, `kube-rbac-proxy` |

Use the `oceanseek`, `logservice`, `minio` and `mc` image repositories
in the OceanBase organization. Source defaults select Quay. GHCR and Docker Hub
can be selected for the OceanBase components using their published versions.
The online release lock pins the exact images used by both deployment channels.

Operator, Dashboard and Helper are built by the existing `release-operator.yml`,
`release-dashboard.yml` and `release-helper.yml` tag workflows in ob-operator.
They keep their independent version inputs. The multi-platform Operator includes
`/oceanseek-readiness` for amd64 and arm64. A dedicated readiness image, when
supplied, must carry that executable. Push an `oceanseek-bootstrap-X.Y.Z` tag on the approved commit of
`codex/oceanseek-compose` to run `oceanseek-compose.yml`. It builds
`oceanseek-bootstrap` for both architectures and publishes the version to Quay,
GHCR and Docker Hub. A fork uses
its configured Quay and Docker Hub namespaces and its GitHub owner for GHCR.

The fields have different meanings:

- `reference`: the canonical tag that the delivered configuration uses.
- `image_id`: the SHA-256 of the image configuration, as reported by Docker.
- `platform`: `linux/amd64` or `linux/arm64`, consistent across each lock.
- `channels`: the channels that use this role; shared roles must include both.
- `source_reference` is an optional alias in the publisher's private input lock.
  The builder can check its local image ID. Published locks contain only the
  runtime identity fields. Preserve the input alongside its checksum privately.
- `manifest_digest`, when present: a registry manifest digest, separate from the
  config image ID. The supplied inspection evidence must prove this digest.
- `publication`: publication evidence status. Current cached builds are marked
  `not-verified`; a canonical `oceanbase/...` name does not mean it was published.

Historical archives retain their original names, versions, IDs, and checksums.
**Never retag an old 4.6.2.0 Observer as 4.6.2.1.** Equal-looking names do not make
binaries equal. An alias may be added only after verifying that its config ID is
exactly the ID approved for the destination reference. The old QA LogService and
the canonical LogService can be aliases because their recorded image IDs match.

## Prepare the image evidence

Use a build host with Python 3.8+, PyYAML, Helm 3, and the Docker Compose CLI.
The current CI uses Python 3.11, PyYAML 6.0.3, and Helm 3.16.4. Building the
configuration bundles does not require a Kubernetes connection.

First obtain the reviewed images from the build cache, a verified archive, or a
verified registry. Ensure all canonical tags are present. The packager does not
create aliases or repair mismatches automatically.

On the host holding the actual images, capture an inspection snapshot:

```bash
python3 - <<'PY'
import json, pathlib, subprocess
lock = json.loads(pathlib.Path('release-images.lock.json').read_text())
refs = [item['reference'] for item in lock['images']]
result = subprocess.run(['docker', 'image', 'inspect', *refs], check=True,
                        stdout=subprocess.PIPE)
pathlib.Path('docker-inspect.json').write_bytes(result.stdout)
PY
```

A full inspection can include aliases. Conflicting identities under the same
reference, a wrong architecture, a source alias with different content, or a
missing canonical tag will fail validation. The snapshot must come from actual
Docker inspection; do not manufacture it from the expected lock for a release.

## Build both configuration bundles

For a configuration-only test preview without available image evidence, explicitly
use `--configuration-only` instead of an evidence argument. This mode still lints,
renders and checks image references, but records zero inspected images and must
not be presented as a verified image release or proof of registry availability.
Both bundles include `tools/package-offline.py` and `OFFLINE.md`: list the actual
configured images, check local Docker identities, or export an offline directory.
The offline exporter includes CR and diagnostic-job images, verifies config IDs
and architecture, and never changes the user's cluster. Its unit/regression tests
are in `test_package_offline.py` (the cached BusyBox smoke test never pulls images).

From the reviewed source checkout, choose a new output directory and version:

```bash
python3 scripts/oceanseek/build-native-packages.py \
  --public-release \
  --source . \
  --lock release-images.lock.json \
  --output /tmp/oceanseek-native-release \
  --version YYYYMMDD-COMMIT \
  --docker-metadata /absolute/path/docker-inspect.json
```

If the build host holds the canonical images, omit `--docker-metadata`; the
builder then runs live `docker image inspect` itself. Alternatively supply one or
more `--image-archive /absolute/path/images.tar.gz` arguments. Archive mode reads
actual Docker-save config bytes and computes their IDs. It verifies archive
checksums but does not claim to have loaded or run the images, verified every
image layer, or tested a registry pull.

The two output archives are:

```text
oceanseek-compose-YYYYMMDD-COMMIT.tar.gz
oceanseek-native-helm-YYYYMMDD-COMMIT.tar.gz
```

They are **configuration bundles, without container images**. Arrange image
import or a verified reachable registry separately before user installation.
Both include the complete release lock, their own channel image inventory,
checksums, and a build validation report. The report states which evidence was
checked and records the source commit and whether the source worktree was dirty.
Use a committed, reviewed source tree for an accepted release. Tar member times use
the source commit timestamp, or an explicit `SOURCE_DATE_EPOCH` between 2000-01-01
and the current time. They do not use epoch zero, avoiding extraction warnings.

For an internal offline delivery, export one shared Docker image archive from
the same validated host. Read the references from the lock rather than maintain
a separate `docker save` list:

```bash
python3 - <<'PY'
import json, pathlib, subprocess
lock = json.loads(pathlib.Path('release-images.lock.json').read_text())
refs = [item['reference'] for item in lock['images']]
subprocess.run(['docker', 'save', '-o', 'oceanseek-images.tar', *refs], check=True)
PY
gzip oceanseek-images.tar
```

Pass this archive to the builder using `--image-archive` to verify its actual
config bytes against the release lock before delivery. Distribute it alongside
both configuration bundles under the same delivery version. A single import
provides both channels' images when they share the same Docker runtime. Never
overwrite an existing released image archive in place.

The generated Compose `.env` contains the canonical tags; Compose retains its
`docker compose up -d` entry point. A configured `IMAGE_REGISTRY` can redirect
those same component names and tags, provided the destination holds identical
images. A per-image environment override is an intentional user override, not
part of the validated default release.

The Helm bundle contains the native OceanSeek and Operator `.tgz` charts,
`image-values.yaml`, `operator-values.yaml`, and single-node/three-node profiles.
Chart defaults and the supplied image values come from the same lock. It contains
**no `install.sh`**: install with native Helm commands described in `USAGE.txt`.
`QUICKSTART.md` gives the default one-command installation with Dashboard enabled;
neither an image override file nor a separate Dashboard release is required.
Operator, cert-manager, worker storage and image availability remain prerequisites.

Both bundles include `registry-mirrors/` from `deploy/registry-mirrors/`. Its
Docker and K3s examples preserve canonical image references and are optional:
neither packaging nor Helm modifies host runtime configuration. Set each mirror endpoint to a trusted URL supplied for that registry. Validate
each locked image from the worker network. Tests are excluded from the
delivered helper directory, and every included file is covered by `SHA256SUMS`.

## Validation gates and acceptance scope

Before creating the delivery archives, the builder:

1. Rejects duplicate roles, wrong component repository names, floating tags,
   old `oceanbase-qa/` delivery aliases, missing roles, and platform differences.
2. Checks the exact config IDs against live inspection, a supplied inspection
   snapshot, or config bytes from Docker-save archives. Metadata-only checks are
   explicitly reported as metadata-only checks.
3. Runs `docker compose config` on the generated bundle and `helm template` on
   the packaged charts for defaults, single-node, and three-node configurations.
4. Checks images by service/container/CR role. Swapping MinIO and mc, or Observer
   and LogService, cannot pass merely because the image set is unchanged.
5. Includes Dashboard diagnostic-job images and Operator kube-rbac-proxy.
   Unknown rendered runtime images are rejected.

Helm and Compose lock the OBAgent image used for database performance metrics.
Helm runs one sidecar per Observer. Compose connects OBAgent to the Observer
by service name and scrapes its independent endpoint.
The Dashboard SQL analyzer is an optional inactive ConfigMap default, explicitly
excluded from this release's validated features; this exception does not allow
an unlocked runtime Pod. Kubernetes, StorageClass, and cert-manager are separately
prepared prerequisites and are not bundled here.

Run the local regressions:

```bash
python3 -m unittest discover -s scripts/oceanseek -p test_native_packages.py -v
go test -count=1 ./charts/oceanseek \
  -run '^TestSharedDatabaseImagesMatchCompose$'
```

The Python packaging test uses **synthetic metadata fixtures** to exercise the
packager, while actually running Helm packaging/rendering and Compose parsing.
It does not validate the availability or runtime behavior of the release images.
CI likewise does not log in to registries, publish images, or operate a cluster.

After a real build, separately test the same locked images in Compose and Helm:
authenticated SQL, restart/data retention, and Dashboard log gathering/download
where applicable. In particular, a shared Observer must satisfy both Compose's
Observer entry point and the Operator's `oceanbase-helper start` contract, plus
the diagnostic Python/archive transport requirements. Record that acceptance
separately; a successful package build is not a successful live installation.
