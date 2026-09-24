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

The script creates `.env` and random MinIO credentials on the first run, builds
the bootstrap utility, starts the services, and waits for OceanBase bootstrap.
Edit `.env` before the first start to change images, ports, resource sizes, or
the private Compose subnet.

Connect to OceanBase with an empty root password:

```bash
obclient -h127.0.0.1 -P32881 -uroot
```

Run a basic persistence check:

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
