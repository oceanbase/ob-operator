# OceanSeek: Docker Compose quickstart

Use this single-machine environment to try OceanSeek SQL without Kubernetes.
It starts eight persistent services: one Observer with its RPM-matched obshell,
three LogServices, MinIO, one OBAgent sidecar, Prometheus and Alertmanager.
Five initialization containers detect the host access IP and create credentials,
buckets and the database;
`Exited (0)` is their expected final state.

The Kubernetes Dashboard is not included in this Compose environment. Use the
Helm deployment for Dashboard and Kubernetes management. This quickstart is for
functional evaluation; three LogServices on one machine do not provide host
fault tolerance.

New installations use the Compose project name `oceanseek`. Existing installations must
keep their original `COMPOSE_PROJECT_NAME` or `-p` value to reuse the same data volumes.
Successful bootstrap prints `OceanSeek deployment completed.` and the actual project name.

## What to prepare

- Linux AMD64, Docker Engine 24+ and the Docker Compose v2 plugin (`docker compose`).
- At least 12 CPU cores, 32 GiB **available** memory and 250 GiB free space in
  Docker's data directory. Leave additional space for your data and logs.
- The eight versioned images selected in `compose.yaml` or the release's `.env`:
  Observer, LogService, MinIO, mc, `oceanseek-bootstrap`, OBAgent, Prometheus and
  Alertmanager. Obtain these from the
  internal release owner and load them into Docker, or use the registry URLs
  supplied with that release. obshell comes from the same `oceanseek` RPM
  as Observer. OBAgent is a separate component image and has its own version.

A prepared release includes the prebuilt bootstrap image or its accessible
registry reference. End users do not build Go code, install Helm or run a
credential-preparation script. The source defaults are versioned references;
a public bootstrap image must be published before its Quay default is usable.

Before installation, run `docker compose version` using the same user and sudo
prefix as deployment. If Docker reports that `compose` is not a command, install
the Docker Compose plugin using the
[official installation instructions](https://docs.docker.com/compose/install/linux/).
When deploying with `sudo`, install the plugin where root can access it and
verify with `sudo docker compose version`.

The Observer entrypoint raises its inherited soft file-descriptor limit to the
container's hard limit before starting Observer and obshell. This also applies
after recreation or restart. A Docker service may otherwise supply a soft limit
of only 1024, causing database initialization to fail with
`OB_TOO_MANY_OPEN_FILES`. Check the actual process limit with
`sudo docker compose exec observer cat /proc/1/limits`; this startup adjustment
does not change the host's limits or increase a low container hard limit.

## 部署前：端口清单与冲突处理

**部署前端口清单（均为 TCP）**

| 用途 | 宿主机端口 → 容器端口 | 默认开放情况 | 修改配置 |
| --- | --- | --- | --- |
| OceanBase SQL | 32881 → 2881 | 默认绑定 0.0.0.0 | `OBSERVER_SQL_HOST_PORT` |
| obshell 管理页面/API | 32886 → 2886 | 默认绑定 0.0.0.0 | `OBSHELL_HOST_PORT` |
| Prometheus | 39090 → 9090 | 仅启用监控端口覆盖文件时开放，默认绑定 127.0.0.1 | `PROMETHEUS_HOST_PORT` |
| Alertmanager | 39093 → 9093 | 仅启用监控端口覆盖文件时开放，默认绑定 127.0.0.1 | `ALERTMANAGER_HOST_PORT` |

默认安装只占用宿主机 **32881、32886**。Observer RPC 2882、LogService 50051/50052、MinIO 9000、OBAgent 8088 等用于容器网络内部通信，不需要在宿主机预留同号端口。使用额外调试覆盖文件时，按其实际映射另行检查。

在部署主机上检查默认端口，并查看已有容器的端口映射：

```bash
sudo ss -ltnp '( sport = :32881 or sport = :32886 )'
sudo docker ps --format 'table {{.Names}}\t{{.Ports}}'
```

`ss` 若出现 LISTEN 记录，先确认占用者；Docker 端口可能通过转发规则实现，也要检查 `docker ps`。启用监控覆盖文件时，再检查 39090、39093。使用自定义端口时检查替换后的值；检查结果仅代表当时状态。

**遇到 `bind: address already in use`：** 例如 `0.0.0.0:32886` 表示宿主机 obshell 端口已被占用，不是数据库内部 2886 冲突。可以为本次部署选择其他空闲端口，不要直接停止未知服务。

在 `compose.yaml` 同目录的 `.env` 中新增或修改下面两项；已有文件请保留其他配置，同名项只保留一条：

```dotenv
OBSERVER_SQL_HOST_PORT=33881
OBSHELL_HOST_PORT=33886
```

33881、33886 仅是示例，使用前同样检查是否空闲。然后在原目录、使用原项目名重试：

```bash
sudo ss -ltnp '( sport = :33881 or sport = :33886 )'
sudo docker compose config --quiet
sudo docker compose up -d && sudo docker compose logs -f observer-bootstrap
```

配置解析检查不检查端口占用。已用 `--env-file` 时修改原环境文件，后续所有 Compose 命令沿用同一 `--env-file`、`-f`、`-p` 参数；显式环境文件场景不要只修改默认 `.env`。端口冲突发生后可能已有部分容器创建，直接用原项目重试即可，**不需要删除数据卷或执行 `down -v`**。最终访问使用新端口，并向客户端来源放行相应防火墙/安全组规则。若创建第二套实例，还需独立项目名和不冲突的容器网段。

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
docker load -i /path/to/oceanseek-images-<delivery-version>.tar.gz
cd /path/to/oceanseek-compose-<delivery-version>
docker compose up -d --pull never
```

The source and licenses for the MinIO server and client are listed in
`scripts/oceanseek/THIRD-PARTY-NOTICES.md` in the repository and in
`THIRD-PARTY-NOTICES.md` at the root of the Compose bundle.

The lightweight Compose and native Helm archives do not contain image layers.
An older Helm image archive can contain different component versions and is not
a substitute for the matching release image archive. Do not retag a different
Observer build to make it appear to match. Release maintainers build both
channels together with `scripts/oceanseek/build-native-packages.py`; see
`scripts/oceanseek/NATIVE-RELEASE.md` in the source repository.

## Keep image names and use a mirror (optional)

The delivery includes `registry-mirrors/README.md`, a Docker configuration example
and a Python 3 merge helper. Preview the optional Docker Hub mirror settings,
then explicitly apply them if suitable for your network:

```bash
sudo python3 registry-mirrors/configure-docker.py --mirror "$DOCKER_HUB_MIRROR"
sudo python3 registry-mirrors/configure-docker.py --mirror "$DOCKER_HUB_MIRROR" --apply
```

Follow that guide to validate and reload Docker, then run `docker compose pull`.
Existing daemon settings and mirrors are preserved. Image names and release tags
stay unchanged. This configuration only accelerates Docker Hub; Quay/GHCR images
still require access to their own registries or the supplied offline archive.
Set DOCKER_HUB_MIRROR to your administrator’s mirror URL before running the preview.
In the source tree, these files are under `deploy/registry-mirrors/`.

## Choose an image registry (optional)

OceanBase images default to `quay.io/oceanbase`. Select `ghcr.io/oceanbase`
or `docker.io/oceanbase` through `OCEANBASE_IMAGE_REGISTRY`. This keeps
Prometheus and Alertmanager in their upstream namespace. Use a complete
`BOOTSTRAP_IMAGE` reference when your fork publishes bootstrap under its own
organization. Check the release's available platforms before installation.

```bash
export OCEANBASE_IMAGE_REGISTRY=ghcr.io/oceanbase
docker compose config --images
docker compose up -d
```

If the default registries are unreachable, use your organization's registry by
setting one variable. Replace the example with a registry/namespace that contains
all eight release images; omit `https://` and any trailing slash. For a private
registry, first run `docker login registry.example.com:5000`.

```bash
export IMAGE_REGISTRY=registry.example.com:5000/team/oceanseek
docker compose config --images
docker compose pull
docker compose up -d
```

You can also prefix each command with the variable instead of exporting it:

```bash
IMAGE_REGISTRY=registry.example.com:5000/team/oceanseek docker compose config --images
IMAGE_REGISTRY=registry.example.com:5000/team/oceanseek docker compose pull
IMAGE_REGISTRY=registry.example.com:5000/team/oceanseek docker compose up -d
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
| `BOOTSTRAP_IMAGE` | `oceanseek-bootstrap:<release-tag>` |
| `MINIO_IMAGE` | `minio:<release-tag>` |
| `MINIO_MC_IMAGE` | `minio-mc:<release-tag>` |
| `OBAGENT_IMAGE` | `quay.io/oceanbase/obagent:4.2.6-100000152026051118` |
| `PROMETHEUS_IMAGE` | `prometheus:<release-tag>` |
| `ALERTMANAGER_IMAGE` | `alertmanager:<release-tag>` |

For persistent configuration, edit `IMAGE_REGISTRY` in the delivered `.env`.
If no `.env` was delivered, copy `.env.example` to `.env` first. **Retain the
delivered image tags/build references**: a release may pin different builds from
the source example, especially `oceanseek-bootstrap`. Do not replace the
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
address and configured SQL port, plus a complete password command referencing
the bootstrap container. Run that command on the Docker host from any directory;
it works after bootstrap exits successfully and does not depend on `-p`, `-f`,
or `--env-file`. Keep the bootstrap container until you retrieve the password.
After recreating containers, copy the command from the latest completion log.
Do not share its output. For the default project, a running Observer also allows:

```bash
sudo docker exec "$(sudo docker ps -q --filter 'label=com.docker.compose.project=oceanseek' --filter 'label=com.docker.compose.service=observer' --filter 'label=com.docker.compose.oneoff=False')" cat /run/secrets/ob_root_password
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
INSERT INTO hello VALUES (1, 'Hello, OceanSeek');
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
OBAgent connects to `observer:2881` on the Compose network. Prometheus
scrapes `obagent:8088`, which keeps its own endpoint across Observer restarts. MinIO and LogService management ports remain internal.

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
| `OBSERVER_CPUS` / `OBSERVER_CPU_COUNT` | `3` / `3` | Container CPU limit / database CPU count |
| `OBSERVER_MEMORY_LIMIT` / `OBSERVER_CONTAINER_MEMORY` | `9G` / `10g` | Database memory target / container limit |
| `LOGSERVICE_CPUS` | `2` | CPU limit of each LogService |
| `LOGSERVICE_MEMORY_LIMIT` / `LOGSERVICE_CONTAINER_MEMORY` | `3686M` / `4g` | Each LogService memory target / container limit |
| `OBAGENT_CPUS` / `OBAGENT_MEMORY` | `0.5` / `512m` | OBAgent container limit |
| `OBSERVER_DATAFILE_SIZE` / `LOGSERVICE_LOG_DISK_SIZE` | `10G` / `9G` | Initial local data / each LogService log space |
| `OBSERVER_DATAFILE_MAXSIZE` / `OBSERVER_DATAFILE_NEXT` | `38G` / `4G` | Observer local data growth ceiling / increment; reserve at least 40 GiB for this volume |
| `COMPOSE_PROJECT_NAME` | `oceanseek` | Groups this installation's containers and volumes |

These are fixed experience defaults, not automatic maximum-use sizing. Keep
container memory above the corresponding application memory target. Initial
datafile and log sizes do not cap total Docker disk use; MinIO data and logs
also consume space.

The test deployment uses these component allocations:
Observer 3 CPU/10 GiB, and three LogServices at 2 CPU/4 GiB each. Compose
intentionally omits ODP and connects directly to Observer. Program memory is
about 90% of each container allocation, matching Helm's default memory policy.
The draft's G/GB are represented as binary GiB here; this is not an exact decimal
GB capacity claim. Reserve SSD space: Observer 40 GiB data + 10 GiB logs, each
LogService 10 GiB store + 10 GiB diagnostic logs. Named volumes do not enforce
these quotas. MinIO data, monitoring, images and OS space are additional; the
250 GiB host recommendation includes a 100 GiB initial MinIO data budget.
The capped components total 9.5 CPU/23.5 GiB, excluding uncapped monitoring and
initialization. Three LogService containers on one host are an experience setup,
not the draft's three independent LogService nodes. Use Helm's
`values-test-environment.yaml` for separate workers.

The Observer starts with a 10 GiB data file and can grow in 4 GiB increments
up to 38 GiB, matching the usable budget of the Helm 40 GiB data volume.
A fixed 10 GiB ceiling can prevent creating a business tenant because the
initial sys unit already requests 10 GiB of data space. The business resource
pool memory minimum is 2 GiB; this is a tenant minimum, not the memory required
by the complete deployment.

These defaults apply to fresh installations. Preserve the original `.env` and
resource settings when upgrading an existing instance: lowering the container
memory to 10g while persisted `memory_limit` remains 10G removes its headroom.
Changing `.env` does not replace persisted Observer configuration. Review current
parameters, assigned tenant resources and disk usage before any resize; do not
automatically shrink existing data files or log budgets to these new defaults.

For basic tenant validation, start with a 1 CPU / 2 GiB memory unit, at least
3 GiB log space and 2 GiB data space, after checking remaining server capacity.
This is a small functional-test configuration, not a production sizing claim.
Keep the whole Compose host budget above: 12 CPU cores, 32 GiB available memory
and 250 GiB free disk; allow more disk for growing MinIO data and logs.

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
`COMPOSE_SUBNET` and all five `*_IP` values together. Existing installations must keep the subnet and service addresses from their
current network. Before replacing a Compose file that used implicit defaults,
record the subnet with `docker network inspect NETWORK` and each service IP
with `docker inspect CONTAINER`, then set those six values in `.env`.
The new-install baseline uses an isolated benchmarking subnet. Changing `.env` resource
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
docker build -f deploy/compose/oceanseek/Dockerfile.bootstrap \
  -t quay.io/oceanbase/oceanseek-bootstrap:0.1.0 .
```

To publish Bootstrap from the release branch, set `BOOTSTRAP_VERSION` to the
approved `X.Y.Z` version and tag its intended commit. The following commands
publish the current remote release-branch tip when the tag is pushed.

```bash
git fetch origin refs/heads/codex/oceanseek-compose
git tag "oceanseek-bootstrap-${BOOTSTRAP_VERSION:?set the approved version}" FETCH_HEAD
git push origin "refs/tags/oceanseek-bootstrap-${BOOTSTRAP_VERSION}"
```

The `oceanseek compose` workflow validates Compose and publishes the amd64 and
arm64 Bootstrap image to Quay, GHCR and Docker Hub. This tag triggers only the
Bootstrap image release. The workflow also accepts manual runs with `version`
and `publish` where GitHub has registered it for dispatch. Configure
`QUAY_IO_PUSH_BASE` and `DOCKER_PUSH_BASE` with the publishing organizations.
GHCR uses the repository owner. Set `BOOTSTRAP_IMAGE` to the actual published
reference when distributing a configuration from a fork.

`compose.build.yaml` is a maintainer-only local build override. For lifecycle
acceptance on an isolated Docker host, select a new `oceanseek-qa-*` project
and unused subnet/port, then run:

```bash
COMPOSE_PROJECT_NAME=oceanseek-qa-acceptance ./scripts/acceptance.sh
```

The test checks initialization, authenticated SQL, wrong-password rejection,
repeated bootstrap, service restart, and `down`/`up` persistence. It preserves
the QA volumes for inspection; `down -v` is an explicit separate action.


### Saved credentials and custom Compose arguments

The completion log prints a complete `docker cp ... - | tar -xOf -` command using the bootstrap container's default Docker identity. Run it on the Docker host from any directory, including after bootstrap has exited successfully. Conflicting project names in `-p` and an environment file do not affect this command. Do not override the bootstrap service's `hostname`; the generated command uses its default Docker container ID. Removing or recreating that container invalidates the old command: copy the latest completion log after `down`/`up`. It reads the saved credential file, not the current database password. Manual password changes do not automatically update that file; use the new password after a change. Do not share the output.

### MinIO console access after installation

There are two separate logins: SSH uses your Docker host machine account and
usual key/password; MinIO uses the saved username/password only in its web form.
The container cannot reliably infer your SSH account, so replace `<ssh-user>`
in the completion message with the account you normally use to SSH to this host.
Do not substitute the MinIO username or the container's root user.

1. On the computer running your browser, open a **new local terminal**. Do not
   run the tunnel from a shell already logged into the Docker host.
2. Run the printed SSH tunnel command, retaining any `-i`, `-J`, or `-p` options
   your normal SSH login needs. `ssh -N` normally stays silent and opens no shell
   after login. This is expected: leave it running; do not press Ctrl+C.
3. Open `http://127.0.0.1:39001/` in the browser on that same computer.
4. In a separate terminal logged into the Docker host, execute the two printed
   `docker cp` commands to read the saved MinIO web username and password. Enter
   them only in the web form. Keep the bootstrap container until retrieval and
   do not share the credential output.
5. When finished, press Ctrl+C in the tunnel terminal to close the connection.

`ExitOnForwardFailure=yes` reports failure to establish the local forward, not
MinIO service readiness. If local port 39001 is busy, select a different first
port in `-L` and use that same port in the browser URL. SSH `Permission denied`
means machine authentication failed; first confirm normal SSH works from the
browser computer. MinIO passwords cannot fix SSH authentication.
The console uses port 9001; port 9000 is the S3 API. No extra installation command
or public port mapping is required. Custom container networks must use the MinIO
address printed by their own deployment, not the default example address.
