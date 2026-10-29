# Native lakehouse release packaging

This workflow prepares the **same release** for Docker Compose and native Helm.
It does not deploy Kubernetes, start a database, pull images, or publish images.
The earlier `build-package.py` and `build-online-package.py` workflows reproduce
historical installer bundles; they are not the native release entry point.

## One component, one image identity

[`release-images.lock.json`](release-images.lock.json) is the release input.
Every component has one role, canonical image reference (including a fixed tag),
full Docker config image ID, and platform. Both delivery formats use that record.
A role cannot select a different Observer build simply because it runs in Helm.

The current internal release contains 13 roles:

| Delivery | Roles |
| --- | --- |
| Shared by Compose and Helm | `observer`, `logservice`, `minio`, `mc` |
| Compose only | `bootstrap` |
| Helm only | `operator`, `dashboard`, `obdiag`, `oceanbase-helper`, `utility`, `prometheus`, `alertmanager`, `kube-rbac-proxy` |

The shared Observer is
`oceanbase/oceanbase-ai-cloud-native:4.6.2.1-100000092026090921.al8`;
LogService uses `oceanbase/oblogservice:1.3.0`. Exact config IDs and the MinIO/mc
versions belong in the lock rather than a second manually maintained list.
The Operator image also supplies the native readiness executable.

The fields have different meanings:

- `reference`: the canonical tag that the delivered configuration uses.
- `image_id`: the SHA-256 of the image configuration, as reported by Docker.
- `platform`: currently `linux/amd64` for every delivered image.
- `channels`: the channels that use this role; shared roles must include both.
- `source_reference`: an existing build-cache alias used for provenance. It does
  not override the canonical reference and is not a promise of public access.
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
lock = json.loads(pathlib.Path('scripts/lakehouse/release-images.lock.json').read_text())
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

From the reviewed source checkout, choose a new output directory and version:

```bash
python3 scripts/lakehouse/build-native-packages.py \
  --source . \
  --lock scripts/lakehouse/release-images.lock.json \
  --output /tmp/lakehouse-native-release \
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
oceanbase-lakehouse-compose-YYYYMMDD-COMMIT.tar.gz
oceanbase-lakehouse-native-helm-YYYYMMDD-COMMIT.tar.gz
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
lock = json.loads(pathlib.Path('scripts/lakehouse/release-images.lock.json').read_text())
refs = [item['reference'] for item in lock['images']]
subprocess.run(['docker', 'save', '-o', 'lakehouse-images.tar', *refs], check=True)
PY
gzip lakehouse-images.tar
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

The Helm bundle contains the native lakehouse and Operator `.tgz` charts,
`image-values.yaml`, `operator-values.yaml`, and single-node/three-node profiles.
Chart defaults and the supplied image values come from the same lock. It contains
**no `install.sh`**: install with native Helm commands described in `USAGE.txt`.

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

The shared-storage release does not configure an OBAgent monitor; the Operator
only creates that container when a MonitorTemplate is configured. Enabling an
unlocked monitor, including a topology override, fails this release gate.
The Dashboard SQL analyzer is an optional inactive ConfigMap default, explicitly
excluded from this release's validated features; this exception does not allow
an unlocked runtime Pod. Kubernetes, StorageClass, and cert-manager are separately
prepared prerequisites and are not bundled here.

Run the local regressions:

```bash
python3 -m unittest discover -s scripts/lakehouse -p test_native_packages.py -v
go test -count=1 ./charts/oceanbase-lakehouse \
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
