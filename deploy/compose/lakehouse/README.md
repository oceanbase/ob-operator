# Lakehouse Docker Compose

This directory starts a single-node OceanBase shared-storage environment with
a three-node OBLogService cluster and MinIO. It is intended for development and
functional testing, not for production or high availability.

## Prerequisites

- Docker Engine 24 or later with Docker Compose v2
- At least 16 CPU cores, 32 GiB available memory, and 60 GiB free disk space
- The images configured in `.env.example` loaded into the local Docker daemon

Prepare credentials and start the environment:

```bash
./scripts/up.sh
```

The script creates `.env` and random MinIO and database root credentials on the
first run, builds the bootstrap utility, starts the services, and verifies
authenticated OceanBase SQL access.
Edit `.env` before the first start to change images, ports, resource sizes, or
the private Compose subnet.

SQL and MinIO ports bind to `127.0.0.1` by default. LogService management and
Observer RPC ports are internal to the Compose network. For remote access,
use SSH forwarding or explicitly configure `HOST_BIND_ADDRESS` and firewall
rules. Extra management ports are available with `compose.debug.yaml`.

Connect to OceanBase with the password from `secrets/ob_root_password`, entering
it at the password prompt rather than including it in shell history:

```bash
obclient -h127.0.0.1 -P32881 -uroot -p
```

Run an automated write/read check (creates only `compose_smoke.t`):

```bash
./scripts/smoke.sh smoke-write
docker compose restart logservice-1 logservice-2 logservice-3 observer
./scripts/smoke.sh smoke-read
docker compose down
./scripts/up.sh
./scripts/smoke.sh smoke-read
```

For the full automated lifecycle acceptance, first configure an isolated
`COMPOSE_PROJECT_NAME=ob-lakehouse-qa-<identifier>`, a separate subnet and unused
ports, then run `./scripts/acceptance.sh`. It deliberately restarts and
recreates that QA project while preserving its volumes. It refuses the default
project and also checks that a wrong password is rejected and credentials do
not change on repeated starts.

Alternatively, run the SQL manually:

```sql
CREATE DATABASE IF NOT EXISTS compose_smoke;
CREATE TABLE IF NOT EXISTS compose_smoke.t (id BIGINT PRIMARY KEY, value VARCHAR(64));
REPLACE INTO compose_smoke.t VALUES (1, 'lakehouse-compose');
SELECT * FROM compose_smoke.t;
```

Restart the long-running services and query the table again:

```bash
docker compose restart logservice-1 logservice-2 logservice-3 observer
```

Stop the environment while preserving data:

```bash
docker compose down
```

Delete the environment and all stored data:

```bash
docker compose down -v
```

Use `docker compose ps -a` and `docker compose logs <service>` for diagnostics.
The fixed container addresses in `.env` are part of the persisted cluster
metadata; changing them after bootstrap requires recreating the volumes.

## Images and offline use

Observer, OBLogService, MinIO and mc versions are explicitly selected in `.env`.
The currently validated Observer/OBLogService combination is 4.6.2.1/1.3.0.
This bundle does not claim 4.8.0 or ARM64 support for those database images.
The sample OBLogService image must be supplied by your trusted distribution.
Do not replace versioned images with `latest` for repeatable installations.

The base `compose.yaml` uses a prebuilt bootstrap image. For development,
`BUILD_BOOTSTRAP=1` adds `compose.build.yaml` and builds it locally. To install
without compiling, load or pull a published bootstrap image, set
`BOOTSTRAP_IMAGE` to its version/digest, and set `BUILD_BOOTSTRAP=0`.
The `lakehouse compose` GitHub workflow validates the bundle and provides a
manual, versioned bootstrap image build/publish action. A public bootstrap
release must be published before using its GHCR tag; the sample does not assume
an unpublished image exists.

For an offline deployment, load all selected images first, including the
prebuilt bootstrap image. Protect and back up the `secrets/` directory along
with the named volumes. Losing the root password file does not reset the
database password; creating a different file does not change existing access.

## Existing empty-password deployments

Stop the original deployment, preserve its `.env`, secrets and volumes, and
upgrade its scripts/configuration. Generate `secrets/ob_root_password` with
`prepare.sh`, then explicitly set `ALLOW_LEGACY_EMPTY_ROOT=true` for the first
upgrade. After a successful authenticated start, set it back to `false`.
The bootstrap tool rejects implicit migration and never resets an existing
non-empty password. Do not use this opt-in against an unrelated database.

## Acceptance and limitations

Startup and initialization retries are bounded and failures return nonzero.
`STARTUP_TIMEOUT_SECONDS` bounds the bootstrap-container wait; individual
services also have healthcheck retry limits and SQL bootstrap has
`BOOTSTRAP_TIMEOUT`. Keep these limits consistent for slow hosts.

This is a single-host functional environment. Its three LogService containers
do not provide host-level fault tolerance. Kubernetes Dashboard/Operator are
not included; their current API requires Kubernetes resources.
