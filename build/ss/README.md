# 从开源 RPM 构建 Operator / Compose 共用业务镜像

同一版本、同一架构的 Observer 镜像只构建一次，Operator 和 Compose 共用。
LogService 也是如此。Observer 与 LogService 是两个组件，各自有一个业务镜像。

构建关系为：

```text
Linux 基础系统 + 当前源码编译的 helper -> oceanbase-helper
oceanbase-helper + oceanbase.ai-ce / libs RPM -> Observer 镜像
oceanbase-helper + oblogservice-ce RPM        -> LogService 镜像
```

两个业务 Dockerfile 都直接 `FROM ${HELPER_IMAGE}`，保留 helper 的运行环境，
不再从 helper 复制一次二进制后切换到另一套基础系统。

## 准备 RPM

本文示例对应 2026-11-02 核对的开源制品：

- `oceanbase.ai-ce`：`4.6.3.0-82026110121.al8.x86_64`，配套同版本的 libs。
- `oblogservice-ce`：`1.3.3-42026110118.al8.x86_64`。

把需要的 RPM 放在以下目录，每个组件只放一个确定版本：

```text
build/ss/rpms/observer/
  oceanbase.ai-ce-4.6.3.0-82026110121.al8.x86_64.rpm
  oceanbase.ai-ce-libs-4.6.3.0-82026110121.al8.x86_64.rpm
  SHA256SUMS
build/ss/rpms/logservice/
  oblogservice-ce-1.3.3-42026110118.al8.x86_64.rpm
  SHA256SUMS
```

对应版本的校验值已写在两个目录的 `SHA256SUMS.example` 中。先和制品平台的
SHA256 核对，再复制为 `SHA256SUMS`：

```bash
cp build/ss/rpms/observer/SHA256SUMS.example build/ss/rpms/observer/SHA256SUMS
cp build/ss/rpms/logservice/SHA256SUMS.example build/ss/rpms/logservice/SHA256SUMS
```

更换 RPM 时，同时更新校验清单、`RPM_VERSION` 和镜像 tag。Dockerfile 不查询制品
平台，也不会在指定版本不存在时换用新版。目录中的 RPM 必须与 SHA256SUMS 清单
一一对应；额外文件、缺文件、校验失败、同包名存在多个版本、
安装后的版本不匹配或动态库缺失都会使构建失败。RPM、二进制和临时下载链接不提交到仓库。

平台当前展示的这批产物标为调试包，下载入口限定内部非生产验证。
这里提供构建方法及内部验证；对外发布时使用获准发布的对应正式制品，替换其校验清单。

## 先准备兼容的 helper

核对的两个主 RPM 都依赖 `GLIBC_2.32`。现有 `build/Dockerfile.obhelper` 使用
Anolis 8.8，不能仅凭“有 helper 二进制”就判断能安装这批 RPM。

`Dockerfile.helper` 使用已固定 digest 的 Alibaba Cloud Linux 3 基础系统，安装
helper 启动需要的工具和 Dashboard 日志下载需要的 Python 3。它只用来构建共享
基础镜像，两个业务镜像再继承它。
已有满足依赖、包含共享存储命令的固定版本 helper 镜像时，可直接复用，无须重建。

以下命令全部从仓库根目录运行，当前样例为 `linux/amd64`，需要 Go 1.25，以及支持 `RUN --mount` 的 Docker BuildKit/buildx：

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -buildvcs=false \
  -o build/ss/bin/oceanbase-helper ./cmd/oceanbase-helper

HELPER_REVISION=$(git rev-parse --short=12 HEAD)
HELPER_IMAGE="local/oceanbase-helper:ss-ce-${HELPER_REVISION}"
docker build --platform linux/amd64 -f build/ss/Dockerfile.helper -t "$HELPER_IMAGE" .
```

新版本 helper 应来自与 Operator 匹配的源码。`BASE_IMAGE` 可以覆盖为管理员提供的
兼容基础系统固定 tag/digest；业务 Dockerfile 的 `HELPER_IMAGE` 也支持 digest。
如基础系统已有镜像，可先导入 Docker；Dockerfile 不添加内网制品仓库或下载凭据。

## 构建两个业务镜像

继续在同一个 shell 运行：

```bash
OBSERVER_IMAGE=local/oceanbase-ai-ce:4.6.3.0-82026110121.al8-r3
LOGSERVICE_IMAGE=local/oblogservice-ce:1.3.3-42026110118.al8-r1

docker build --platform linux/amd64 -f build/ss/Dockerfile.observer \
  --build-arg HELPER_IMAGE="$HELPER_IMAGE" \
  --build-arg RPM_VERSION=4.6.3.0-82026110121.al8 \
  -t "$OBSERVER_IMAGE" .

docker build --platform linux/amd64 -f build/ss/Dockerfile.oblogservice \
  --build-arg HELPER_IMAGE="$HELPER_IMAGE" \
  --build-arg RPM_VERSION=1.3.3-42026110118.al8 \
  -t "$LOGSERVICE_IMAGE" .
```

构建通过 BuildKit 只读 bind mount 安装本地 RPM，校验脚本同样只读挂载；不会先
COPY 安装包到镜像层。RPM 及校验脚本不会残留在最终镜像层中。构建会保留 RPM
自带的程序和配置目录。LogService 仅补上 Operator
要求的 helper 链接：`/home/admin/oblogservice/bin/oceanbase-helper`。
Observer 当前以 root 启动，镜像构建时将本地 `etc` 目录属主设为 root，满足程序的
启动 UID 校验；不递归更改安装目录或已有数据卷的属主。
不在 Dockerfile 中创建或初始化 `store`、持久化 `log`，也不声明匿名 `VOLUME`。
RPM 本身包含的空目录可能仍会存在；实际数据目录在部署时挂载 PVC 或 Compose 卷，
由 helper / Compose 启动脚本准备。卷权限按部署配置处理，不对整个 `/home/admin`
递归更改属主，也不使用 `useradd ... || true` 掩盖构建错误。

当前 Observer RPM 已包含 `/home/admin/oceanbase/bin/obshell`，因此 obshell 与
Observer 使用同一个业务镜像和同一个 RPM 版本。构建会检查该文件存在，Compose
启动 Observer 后同时启动 obshell agent；无需再构建一个 obshell 镜像。

OBAgent 保持为独立镜像。它的 OceanBase 采集配置连接 `127.0.0.1:2881`，所以
Compose 让 OBAgent 与 Observer 共用网络命名空间，效果与 Operator 在同一个 Pod
中增加 OBAgent sidecar 一致。这样能访问本机 Observer，同时保留独立升级、资源限制
和健康检查能力。不要把 OBAgent RPM 安装进 Observer 镜像。

Compose 固定使用 `oceanbase/obagent:4.2.4-200000022025090416`、
`quay.io/prometheus/prometheus:v2.51.1` 和
`quay.io/prometheus/alertmanager:v0.27.0`。Prometheus 和 Alertmanager 的版本不与
OceanBase RPM 共用版本号；交付版本锁固定实际验证的组合，避免使用浮动 tag。

## 两种部署使用同一镜像

- Operator 的 Observer 启动命令：`/home/admin/oceanbase/bin/oceanbase-helper start`。
- Operator 的 LogService 启动命令：`/home/admin/oblogservice/bin/oceanbase-helper logservice start`。
- Compose 覆盖入口为交付包挂载的 `observer-entrypoint.sh` / `logservice-entrypoint.sh`。

Dockerfile 使用 `CMD` 并清空继承的 `ENTRYPOINT`，允许两种部署覆盖启动方式。
Observer 默认工作目录为 `/home/admin/oceanbase`，LogService 为 `/home/admin/oblogservice`。

镜像初步检查：

```bash
docker run --rm --entrypoint /home/admin/oceanbase/bin/observer "$OBSERVER_IMAGE" -V
docker run --rm --entrypoint /bin/bash "$LOGSERVICE_IMAGE" \
  -c 'test -x /home/admin/oblogservice/bin/oblogservice && test -x /home/admin/oblogservice/bin/oceanbase-helper'
```

当前源码的 helper 初始化路径会读取 Kubernetes 配置，因此镜像构建阶段不执行
helper，也不向镜像内写入 kubeconfig。Operator 在 Pod 中提供集群访问配置；
Compose 使用自己的启动脚本，不执行这个 helper。新 RPM 的完整启动和集群初始化
仍需用对应的部署配置验证，不能用一次版本检查代替。

在 Compose 的 `.env` 中设置 `OBSERVER_IMAGE` 和 `LOGSERVICE_IMAGE`；Helm 中设置
`observer.image` 和 `logService.image` 为相同的两个引用。这里不自动修改先前版本的
`release-images.lock.json`，不把新 RPM 重标成旧版本镜像，也不执行镜像推送。

## 关于先删除 LogService

当前源码会检查同命名空间的 OBCluster 是否还通过 `logServiceRef.name` 引用它。
若仍被引用，删除请求会留下 `deletionTimestamp`，但控制器阻止移除 finalizer，
并记录 `DeletionBlocked` 事件。因此先删 LogService 不会正常完成删除；应先完成
OBCluster 的删除，再删除 LogService。此行为来自控制器，与镜像 Dockerfile 无关。

## Compose 内置运行脚本的派生镜像

需要将启动脚本内置到 `/opt/lakehouse/` 时，使用 `Dockerfile.compose-runtime`。
准备独立构建上下文，将已审定版本的脚本置于 `lakehouse/`，并将
`OBSERVER_DIGEST_REFERENCE` 设为已发布的 `仓库地址@sha256:摘要`，然后执行：

```bash
docker build -f build/ss/Dockerfile.compose-runtime \
  --build-arg BASE_IMAGE="$OBSERVER_DIGEST_REFERENCE" \
  -t registry.example.com/observer:compose-version /path/to/runtime-context
```

LogService 派生镜像使用同一配方并传入对应的 LogService digest。
`BASE_IMAGE` 无默认浮动 tag，并拒绝不含 digest 的输入。发布脚本应从推送结果或锁文件
读取该引用，同时把它记入交付清单；不要改写历史版本的构建快照。

## LogService process supervision

The final LogService image installs and checks `procps-ng` explicitly, even when
`HELPER_IMAGE` is supplied by another build. Do not assume a helper binary brings
its original runtime tools with it. The native helper and Compose launcher inspect
Linux `/proc` directly, so missing `pgrep` cannot become a false startup timeout.
Process inspection errors and a missing/exited daemon are distinct conditions.

Before publishing, run the helper process regression in an isolated Linux PID
namespace, then run both real startup paths with `pgrep` temporarily hidden in
the test containers. Verify RPC/HTTP access, zero restarts beyond the Compose
300-second boundary, and exit on real daemon failure. For the native helper,
setting `LOGSERVICE_START_TIMEOUT_SECONDS=10` exercises its timeout boundary
without waiting for the default 1800 seconds. Also check graceful termination.
An initial open port or a successful bootstrap alone does not establish that the
container will survive its startup deadline. Run these checks on pooled test
machines, not development/build hosts.
