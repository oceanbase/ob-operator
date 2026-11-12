# OceanBase lakehouse: Docker Compose quickstart

Use this single-machine environment to try lakehouse SQL without Kubernetes.
It starts eight persistent services: one Observer with its RPM-matched obshell,
three LogServices, MinIO, one OBAgent sidecar, Prometheus and Alertmanager.
Five initialization containers detect the host access IP and create credentials,
buckets and the database;
`Exited (0)` is their expected final state.

The Kubernetes Dashboard is not included in this Compose environment. Use the
Helm deployment for Dashboard and Kubernetes management. This quickstart is for
functional evaluation; three LogServices on one machine do not provide host
fault tolerance.

## What to prepare

- Linux AMD64, Docker Engine 24+ and the Docker Compose v2 plugin (`docker compose`).
- At least 8 CPU cores, 32 GiB **available** memory and 100 GiB free space in
  Docker's data directory. Leave additional space for your data and logs.
- The eight versioned images selected in `compose.yaml` or the release's `.env`:
  Observer, LogService, MinIO, mc, `lakehouse-bootstrap`, OBAgent, Prometheus and
  Alertmanager. Obtain these from the
  internal release owner and load them into Docker, or use the registry URLs
  supplied with that release. obshell comes from the same `oceanbase.ai-ce` RPM
  as Observer. OBAgent is a separate component image and has its own version.

A prepared release includes the prebuilt bootstrap image or its accessible
registry reference. End users do not build Go code, install Helm or run a
credential-preparation script. The source defaults are versioned references;
a public bootstrap image must be published before its GHCR default is usable.

Before installation, run `docker compose version` using the same user and sudo
prefix as deployment. If Docker reports that `compose` is not a command, install
the Docker Compose plugin using the
[official installation instructions](https://docs.docker.com/compose/install/linux/).
When deploying with `sudo`, install the plugin where root can access it and
verify with `sudo docker compose version`.

## Matching Helm and Compose releases

Use the Compose bundle and images from the same delivery version. Both native
Helm and Compose bundles are generated from `release-images.lock.json`: their
Observer, LogService, MinIO, mc, Prometheus and Alertmanager references, versions
and config image IDs are identical. Helm includes additional Kubernetes
components; Compose additionally uses the bootstrap and OBAgent images.
`image-lock.json` lists the subset needed by this
bundle. Config image IDs are not registry manifest digests.

For an internal offline delivery, load the matching shared image archive once
into Docker, then start from the extracted **Compose** directory:

```bash
docker load -i /path/to/oceanbase-lakehouse-images-<delivery-version>.tar.gz
cd /path/to/oceanbase-lakehouse-compose-<delivery-version>
docker compose up -d --pull never
```

The matching Compose image archive includes the pinned open-source
`minio/minio` and `minio/mc` images. No AIStor image or license is required.

The lightweight Compose and native Helm archives do not contain image layers.
An older Helm image archive can contain different component versions and is not
a substitute for the matching release image archive. Do not retag a different
Observer build to make it appear to match. Release maintainers build both
channels together with `scripts/lakehouse/build-native-packages.py`; see
`scripts/lakehouse/NATIVE-RELEASE.md` in the source repository.

## Keep image names and use a mirror (optional)

The delivery includes `registry-mirrors/README.md`, a Docker configuration example
and a Python 3 merge helper. Preview the optional Docker Hub mirror settings,
then explicitly apply them if suitable for your network:

```bash
sudo python3 registry-mirrors/configure-docker.py
sudo python3 registry-mirrors/configure-docker.py --apply
```

Follow that guide to validate and reload Docker, then run `docker compose pull`.
Existing daemon settings and mirrors are preserved. Image names and release tags
stay unchanged. This configuration only accelerates Docker Hub; Quay/GHCR images
still require access to their own registries or the supplied offline archive.
The configuration includes the complete 14-address mirror list. Inclusion does
not guarantee availability of release images at every endpoint.
In the source tree, these files are under `deploy/registry-mirrors/`.

## Choose an image registry (optional)

If the default registries are unreachable, use your organization's registry by
setting one variable. Replace the example with a registry/namespace that contains
all eight release images; omit `https://` and any trailing slash. For a private
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
| `OBAGENT_IMAGE` | `obagent:<release-tag>` |
| `PROMETHEUS_IMAGE` | `prometheus:<release-tag>` |
| `ALERTMANAGER_IMAGE` | `alertmanager:<release-tag>` |

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
docker compose up -d && docker compose logs -f observer-bootstrap
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

The final `observer-bootstrap` log prints the password retrieval command and
the SQL, obshell Dashboard, Prometheus and Alertmanager addresses using the
automatically detected Docker host IP.
`docker compose up -d` remains the native start command; the chained `logs -f`
only follows initialization output.

OBAgent starts after the authenticated bootstrap creates its read-only `monitor`
account. Prometheus then waits for OBAgent and Alertmanager. The Observer
entrypoint registers both internal service addresses with obshell through its
local Unix socket, following OBD's registration API. The performance and alert
pages therefore open without asking the user for addresses or credentials.
Verify the chain:

```bash
docker compose ps
docker compose logs --tail=100 obagent prometheus alertmanager
```

## Connect and try SQL

The completion log prints a complete MySQL command with the discovered host
address and configured SQL port, plus a complete password command containing
the actual Compose project name (including a name chosen with `-p`). Run the
password command on the Docker host from any directory. It uses container labels,
so custom `-f` and `--env-file` arguments are not needed. Do not share its output.
For the default project:

```bash
sudo docker exec "$(sudo docker ps -q --filter 'label=com.docker.compose.project=ob-lakehouse' --filter 'label=com.docker.compose.service=observer' --filter 'label=com.docker.compose.oneoff=False')" cat /run/secrets/ob_root_password
```

You may omit `sudo` if your user can access Docker directly.

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

SQL is published on `0.0.0.0:32881` by default. Connect from another machine
with the Docker host address and allow TCP port 32881 in the host firewall or
security group:

```bash
mysql -h<Docker-host-IP> -P32881 -uroot@sys -p
```

Use the host address rather than a container IP. The management endpoints are:

- obshell Dashboard: `http://<Docker-host-IP>:32886/` (published on
  `0.0.0.0:32886` by default; use HTTP, not HTTPS)
Prometheus and Alertmanager remain on the Compose network by default. To
publish them on host loopback, use the monitoring overlay:

```bash
docker compose -f compose.yaml -f compose.monitoring.yaml up -d
```

Use the same two `-f` arguments for subsequent start, upgrade, password, log
and shutdown commands. Prometheus is then at `http://127.0.0.1:39090/` and
Alertmanager at `http://127.0.0.1:39093/` on the Docker host. For remote access,
use SSH forwarding or explicitly set `MONITORING_BIND_ADDRESS` to the desired
interface. Prometheus's HTTP reload API stays disabled; restart the container
to apply monitoring configuration changes. Alertmanager's reload endpoint remains
available inside the Compose network; its host port is unpublished by default.
The loopback overlay does not add authentication. Keep it on loopback or place
an authenticated proxy in front of any remote monitoring access.

The obshell process is the binary shipped by the exact Observer RPM. Each
OBAgent monitors one Observer through `127.0.0.1:2881`; Compose therefore runs
it as a separate container sharing the Observer network namespace, matching the
Operator sidecar model. MinIO and LogService management ports remain internal.

## Optional configuration

The release defaults work without a user-created `.env`. If the release already
includes `.env`, edit that file and retain its image references. Otherwise copy
`.env.example` to `.env` before the first start.

| Setting | Default | Purpose |
| --- | --- | --- |
| `HOST_ACCESS_IP` | auto-detected | Address printed for clients; set explicitly on multi-NIC or NAT hosts |
| `OBSERVER_SQL_HOST_PORT` | `32881` | SQL port on the Docker host |
| `OBSHELL_HOST_PORT` | `32886` | obshell port on the Docker host |
| `PROMETHEUS_HOST_PORT` | `39090` | Prometheus port on the Docker host |
| `ALERTMANAGER_HOST_PORT` | `39093` | Alertmanager port on the Docker host |
| `OBSERVER_SQL_BIND_ADDRESS` | `0.0.0.0` | Host interface accepting SQL connections |
| `OBSHELL_BIND_ADDRESS` | `0.0.0.0` | Host interface accepting obshell Dashboard connections |
| `MONITORING_BIND_ADDRESS` | `127.0.0.1` | Host interface for the optional monitoring overlay |
| `AUTO_REGISTER_MONITORING` | `true` | Register Prometheus and Alertmanager in obshell automatically |
| `PROMETHEUS_INTERNAL_URL` | `http://prometheus:9090/` | Stable Compose address stored in obshell |
| `ALERTMANAGER_INTERNAL_URL` | `http://alertmanager:9093` | Stable Compose address stored in obshell |
| `OBSERVER_CPUS` / `OBSERVER_CPU_COUNT` | `2` / `2` | Container CPU limit / database CPU count |
| `OBSERVER_MEMORY_LIMIT` / `OBSERVER_CONTAINER_MEMORY` | `10G` / `12g` | Database memory target / container limit |
| `LOGSERVICE_CPUS` | `1` | CPU limit of each LogService |
| `LOGSERVICE_MEMORY_LIMIT` / `LOGSERVICE_CONTAINER_MEMORY` | `3G` / `4g` | Each LogService memory target / container limit |
| `OBAGENT_CPUS` / `OBAGENT_MEMORY` | `0.5` / `512m` | OBAgent container limit |
| `OBSERVER_DATAFILE_SIZE` / `LOGSERVICE_LOG_DISK_SIZE` | `10G` / `10G` | Initial local data / each LogService log space |
| `OBSERVER_DATAFILE_MAXSIZE` / `OBSERVER_DATAFILE_NEXT` | `47G` / `5G` | Observer local data growth ceiling / increment; reserve at least 50 GiB for this volume |
| `COMPOSE_PROJECT_NAME` | `ob-lakehouse` | Groups this installation's containers and volumes |

These are fixed experience defaults, not automatic maximum-use sizing. Keep
container memory above the corresponding application memory target. Initial
datafile and log sizes do not cap total Docker disk use; MinIO data and logs
also consume space.

The Observer starts with a 10 GiB data file and can grow in 5 GiB increments
up to 47 GiB, matching the usable budget of the Helm 50 GiB data volume.
A fixed 10 GiB ceiling can prevent creating a business tenant because the
initial sys unit already requests 10 GiB of data space. The business resource
pool memory minimum is 2 GiB; this is a tenant minimum, not the memory required
by the complete deployment.

For an existing installation, changing `.env` alone does not replace the
persisted Observer configuration. After checking free space on the data volume,
apply the matching values through a sys-tenant SQL session:

```sql
ALTER SYSTEM SET __min_full_resource_pool_memory=2147483648;
ALTER SYSTEM SET datafile_maxsize='47G';
ALTER SYSTEM SET datafile_next='5G';
```

Read back the parameters and `oceanbase.GV$OB_SERVERS` before creating a new
resource pool. These changes do not require restarting the database.

For basic tenant validation, start with a 1 CPU / 2 GiB memory unit, at least
3 GiB log space and 2 GiB data space, after checking remaining server capacity.
This is a small functional-test configuration, not a production sizing claim.
Keep the whole Compose host budget above: 8 CPU cores, 32 GiB available memory
and 100 GiB free disk; allow more disk for growing MinIO data and logs.

A new tenant's first major compaction includes its system tablets as well as
the test table. Allow up to 15 minutes for acceptance and record progress from
`GV$OB_COMPACTION_PROGRESS`; a three-minute observation timeout alone does not
prove a resource shortage. Capture the new `FROZEN_SCN` after issuing the freeze,
then require `LAST_SCN >= FROZEN_SCN`, `STATUS='IDLE'` and `IS_ERROR='NO'` in
`CDB_OB_MAJOR_COMPACTION` for that tenant. The initial `SCN=1, IDLE` state before
the asynchronous freeze starts is not a successful major compaction. Finally,
verify the test data on a new SQL connection and check its objects in MinIO.

Cluster IDs, bucket names and static container addresses are persisted in the
database. Configure these before first use. To avoid a subnet collision, change
`COMPOSE_SUBNET` and all five `*_IP` values together. Changing `.env` resource
values on an existing deployment does not necessarily reconfigure persisted
database settings; use this version's defaults for the first experience.

Host IP detection runs once in a short-lived container with host networking and
uses the source address of the default IPv4 route. Set `HOST_ACCESS_IP` in
`.env` when clients must use another interface or a NAT/public address.

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
docker compose restart logservice-1 logservice-2 logservice-3 observer obagent prometheus alertmanager
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
docker compose logs --tail=100 observer obagent prometheus alertmanager logservice-1
```

- **Image pull failed / Docker Hub timeout:** use the optional mirror or registry
  configuration above and check `docker compose config --images`, then retry `docker compose
  pull`. Other services may report `context canceled` after one pull fails; fix
  that pull error first. The registry must contain the exact release images,
  including the prebuilt bootstrap image.
- **Container exited / OOMKilled:** check available host memory and limits. The
  defaults cap each LogService to avoid consuming the host's full memory.
- **Initialization failed:** the dependency chain stops the dependent service.
  Correct the reported issue and run `docker compose up -d` again.
- **Authentication rejected:** verify the credentials volume matches this
  deployment. Initialization never silently replaces a persisted password.
- **OBAgent unavailable:** confirm `observer-bootstrap` completed and inspect
  the OBAgent log. Its monitor password is generated separately from root.
- **Monitoring setup page still appears:** inspect the Observer log for the two
  `obshell ... registered` messages and confirm Prometheus and Alertmanager are
  healthy. Registration is idempotent and runs again when Observer restarts.
- **Wrong address printed:** set `HOST_ACCESS_IP` in `.env` to the address
  reachable from the client, then recreate `host-info-init` and
  `observer-bootstrap`.
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


### Saved credentials and custom Compose arguments

The completion log prints a complete Docker command using the actual Compose project label. Run it on the Docker host from any directory; `-f`, `--env-file`, and custom container names do not need to be reconstructed. It requires exactly one running Observer for that project. It reads the saved credential file, not the current database password. Manual password changes do not automatically update that file; use the new password after a change. Do not share the output.
