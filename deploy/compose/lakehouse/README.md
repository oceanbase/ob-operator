# OceanBase lakehouse: Docker Compose quickstart

Use this single-machine environment to try lakehouse SQL without Kubernetes.
It starts five persistent services: one Observer, three LogServices and MinIO.
Four initialization containers create credentials, buckets and the database;
`Exited (0)` is their expected final state.

The Kubernetes Dashboard is not included in this Compose environment. Use the
Helm deployment for Dashboard and Kubernetes management. This quickstart is for
functional evaluation; three LogServices on one machine do not provide host
fault tolerance.

## What to prepare

- Linux AMD64, Docker Engine 24+ and the Docker Compose v2 plugin (`docker compose`).
- At least 8 CPU cores, 32 GiB **available** memory and 100 GiB free space in
  Docker's data directory. Leave additional space for your data and logs.
- The five versioned images selected in `compose.yaml` or the release's `.env`:
  Observer, LogService, MinIO, mc and `lakehouse-bootstrap`. Obtain these from the
  internal release owner and load them into Docker, or use the registry URLs
  supplied with that release. The current Observer/LogService combination is
  4.6.2.1/1.3.0 and is not asserted to be publicly downloadable or ARM64 compatible.

A prepared release includes the prebuilt bootstrap image or its accessible
registry reference. End users do not build Go code, install Helm or run a
credential-preparation script. The source defaults are versioned references;
a public bootstrap image must be published before its GHCR default is usable.

## Matching Helm and Compose releases

Use the Compose bundle and images from the same delivery version. Both native
Helm and Compose bundles are generated from `release-images.lock.json`: their
Observer, LogService, MinIO and mc references, versions and config image IDs are
identical. Helm includes additional Kubernetes components; Compose additionally
uses the bootstrap image. `image-lock.json` lists the subset needed by this
bundle. Config image IDs are not registry manifest digests.

For an internal offline delivery, load the matching shared image archive once
into Docker, then start from the extracted **Compose** directory:

```bash
docker load -i /path/to/oceanbase-lakehouse-images-<delivery-version>.tar.gz
cd /path/to/oceanbase-lakehouse-compose-<delivery-version>
docker compose up -d --pull never
```

The lightweight Compose and native Helm archives do not contain image layers.
An older Helm image archive can contain different component versions and is not
a substitute for the matching release image archive. Do not retag a different
Observer build to make it appear to match. Release maintainers build both
channels together with `scripts/lakehouse/build-native-packages.py`; see
`scripts/lakehouse/NATIVE-RELEASE.md` in the source repository.

## Choose an image registry (optional)

If the default registries are unreachable, use your organization's registry by
setting one variable. Replace the example with a registry/namespace that contains
all five release images; omit `https://` and any trailing slash. For a private
registry, first run `docker login registry.example.com:5000`.

```bash
export IMAGE_REGISTRY=registry.example.com:5000/team/lakehouse
docker compose config --images
docker compose pull
docker compose up -d
```

You can also prefix each command with the variable instead of exporting it:

```bash
IMAGE_REGISTRY=registry.example.com:5000/team/lakehouse docker compose config --images
IMAGE_REGISTRY=registry.example.com:5000/team/lakehouse docker compose pull
IMAGE_REGISTRY=registry.example.com:5000/team/lakehouse docker compose up -d
```

Inline variables apply only to that command. Use the same setting for later
Compose commands, or save it in `.env`. `docker compose up` has no `--registry`
or `-e` option for selecting images.

`IMAGE_REGISTRY` replaces the registry/namespace and keeps each image name and
release tag:

| Image override | Image under `IMAGE_REGISTRY` |
| --- | --- |
| `OBSERVER_IMAGE` | `oceanbase-ai-cloud-native:<release-tag>` |
| `LOGSERVICE_IMAGE` | `oblogservice:<release-tag>` |
| `BOOTSTRAP_IMAGE` | `lakehouse-bootstrap:<release-tag>` |
| `MINIO_IMAGE` | `minio:<release-tag>` |
| `MINIO_MC_IMAGE` | `mc:<release-tag>` |

For persistent configuration, edit `IMAGE_REGISTRY` in the delivered `.env`.
If no `.env` was delivered, copy `.env.example` to `.env` first. **Retain the
delivered image tags/build references**: a release may pin different builds from
the source example, especially `lakehouse-bootstrap`. Do not replace the
delivered `.env` with the source example.

To keep a separate configuration file, copy the complete delivered `.env` and
edit its `IMAGE_REGISTRY` value; a minimal `--env-file` containing only the
registry would omit the release's image pins:

```bash
cp .env my-registry.env
# Edit IMAGE_REGISTRY in my-registry.env; preserve the other release settings.
docker compose --env-file my-registry.env config --images
docker compose --env-file my-registry.env pull
docker compose --env-file my-registry.env up -d
```

If your registry uses different image names or paths, set any individual
`*_IMAGE` variable from the table to a full `registry/path/name:tag` or
`registry/path/name@sha256:<digest>` reference. These overrides take precedence
over `IMAGE_REGISTRY`; check `docker compose config --images` before pulling.

This setting selects image locations; it does not configure a Docker Engine
registry mirror or upload images. The selected registry must be reachable and
contain the exact compatible release images. There is no automatic fallback to
an unverified public mirror, and the source defaults do not guarantee public
image availability.

## Start

Extract the supplied quickstart archive and enter its directory, then run:

```bash
docker compose up -d
docker compose logs -f observer-bootstrap
```

Initialization is automatic. Wait for:

```text
OceanBase bootstrap completed with authenticated SQL access
```

`up -d` can return while SQL initialization is still running. The Observer
healthcheck checks its TCP listener; it does not by itself prove SQL readiness.
After bootstrap completes, use this command to verify authenticated database
readiness and get a successful exit code:

```bash
docker compose run --rm --no-deps observer-bootstrap --mode check
```

The check is read-only. If initialization fails, inspect the relevant service
logs and rerun `docker compose up -d` after correcting the cause.

## Connect and try SQL

Display the generated database password locally:

```bash
docker compose run --rm --no-deps observer-bootstrap --mode show-password
```

Connect using an existing MySQL-compatible client and enter that password at
the prompt:

```bash
mysql -h127.0.0.1 -P32881 -uroot@sys -p
```

```sql
CREATE DATABASE IF NOT EXISTS demo;
USE demo;
CREATE TABLE IF NOT EXISTS hello (id BIGINT PRIMARY KEY, message VARCHAR(128));
INSERT INTO hello VALUES (1, 'Hello, OceanBase lakehouse');
SELECT * FROM hello;
```

No client installed? Run the bundled SQL write/read check:

```bash
docker compose run --rm --no-deps observer-bootstrap --mode smoke-write
```

This creates only `compose_smoke.t` and verifies a stored row. Later use
`--mode smoke-read` to verify the same row survives restart.

SQL binds to `127.0.0.1:32881` on the Docker host by default. For access from
another machine, forward the port over SSH:

```bash
ssh -N -L 32881:127.0.0.1:32881 user@docker-host
```

Then connect the local client to `127.0.0.1:32881`. Alternatively set
`HOST_BIND_ADDRESS` to the host interface address (or `0.0.0.0`) in `.env` and
allow port 32881 from your client network. Do not use container IPs as the
external connection address. MinIO and LogService management ports remain
internal to the Compose network.

## Optional configuration

The release defaults work without a user-created `.env`. If the release already
includes `.env`, edit that file and retain its image references. Otherwise copy
`.env.example` to `.env` before the first start.

| Setting | Default | Purpose |
| --- | --- | --- |
| `OBSERVER_SQL_HOST_PORT` | `32881` | SQL port on the Docker host |
| `HOST_BIND_ADDRESS` | `127.0.0.1` | Host interface accepting SQL connections |
| `OBSERVER_CPUS` / `OBSERVER_CPU_COUNT` | `2` / `2` | Container CPU limit / database CPU count |
| `OBSERVER_MEMORY_LIMIT` / `OBSERVER_CONTAINER_MEMORY` | `10G` / `12g` | Database memory target / container limit |
| `LOGSERVICE_CPUS` | `1` | CPU limit of each LogService |
| `LOGSERVICE_MEMORY_LIMIT` / `LOGSERVICE_CONTAINER_MEMORY` | `3G` / `4g` | Each LogService memory target / container limit |
| `OBSERVER_DATAFILE_SIZE` / `LOGSERVICE_LOG_DISK_SIZE` | `10G` / `10G` | Initial local data / each LogService log space |
| `COMPOSE_PROJECT_NAME` | `ob-lakehouse` | Groups this installation's containers and volumes |

These are fixed experience defaults, not automatic maximum-use sizing. Keep
container memory above the corresponding application memory target. Initial
datafile and log sizes do not cap total Docker disk use; MinIO data and logs
also consume space.

Cluster IDs, bucket names and static container addresses are persisted in the
database. Configure these before first use. To avoid a subnet collision, change
`COMPOSE_SUBNET` and all five `*_IP` values together. Changing `.env` resource
values on an existing deployment does not necessarily reconfigure persisted
database settings; use this version's defaults for the first experience.

## Stop, resume and remove

Stop containers and retain all database data and passwords:

```bash
docker compose down
```

Resume with the same directory, project name and configuration:

```bash
docker compose up -d
```

Restart only running database services:

```bash
docker compose restart logservice-1 logservice-2 logservice-3 observer
```

Delete this installation **including all database/object-store data and passwords**:

```bash
docker compose down -v
```

The `credentials` named volume stores generated secrets with private file
permissions. Back it up with the other data volumes. Startup preserves existing
passwords; it does not reset them. Changing the database password directly with
`ALTER USER` does not update the credentials volume; subsequent checks will
fail until the stored credential is updated consistently.

## Troubleshooting

```bash
docker compose ps -a
docker compose logs --tail=100 credentials-init minio-init logservice-bootstrap observer-bootstrap
docker compose logs --tail=100 observer logservice-1
```

- **Image pull failed / Docker Hub timeout:** use the registry configuration
  above and check `docker compose config --images`, then retry `docker compose
  pull`. Other services may report `context canceled` after one pull fails; fix
  that pull error first. The registry must contain the exact release images,
  including the prebuilt bootstrap image.
- **Container exited / OOMKilled:** check available host memory and limits. The
  defaults cap each LogService to avoid consuming the host's full memory.
- **Initialization failed:** the dependency chain stops the dependent service.
  Correct the reported issue and run `docker compose up -d` again.
- **Authentication rejected:** verify the credentials volume matches this
  deployment. Initialization never silently replaces a persisted password.
- **Port/subnet conflict:** choose an unused SQL port and subnet before first use.

Container console logs are rotated (3 files of 10 MiB per container). Database
log files and object data persist in named volumes and need host disk monitoring.

## Maintainer build and acceptance

Build the precompiled bootstrap image during release preparation:

```bash
docker build -f deploy/compose/lakehouse/Dockerfile.bootstrap \
  -t ghcr.io/oceanbase/lakehouse-bootstrap:0.1.0 .
```

The `lakehouse compose` workflow tests the bootstrap utility and validates the
Compose configuration. Its manual publishing action can publish a versioned
bootstrap image to GHCR. Match `BOOTSTRAP_IMAGE` in the distributed configuration
to the actual published or bundled image; do not give end users local build steps.

`compose.build.yaml` is a maintainer-only local build override. For lifecycle
acceptance on an isolated Docker host, select a new `ob-lakehouse-qa-*` project
and unused subnet/port, then run:

```bash
COMPOSE_PROJECT_NAME=ob-lakehouse-qa-acceptance ./scripts/acceptance.sh
```

The test checks initialization, authenticated SQL, wrong-password rejection,
repeated bootstrap, service restart, and `down`/`up` persistence. It preserves
the QA volumes for inspection; `down -v` is an explicit separate action.
