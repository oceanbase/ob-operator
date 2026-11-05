# 在线部署与离线准备

本包不包含镜像层。启动入口保持 `docker compose up -d` 或原生 Helm 命令。
当前测试版的部分自建镜像尚未验证仓库发布，不能保证在全新机器直接在线拉取。
已有匹配镜像的测试机可以直接使用；制作离线包可从已有缓存导出。
不要将旧镜像改名冒充当前版本，也不要改用 AIStor 或 latest 绕过缺失镜像。

## 查看本次部署需要哪些镜像

在解压后的目录中执行。Python 需要 3.8+；Compose 需要 Docker Compose v2；
Helm 需要 Helm 3。仅查看清单不需要 Docker daemon，也不会下载镜像或部署集群。

```bash
python3 -m pip install -r tools/requirements.txt
python3 tools/package-offline.py
```

工具从实际 Compose 配置或 Helm 渲染结果收集镜像，包含初始化/Hook、
OBCluster/OBLogServiceCluster 中的数据库镜像和 Dashboard 动态诊断镜像。
Helm 包同时收集 Operator 镜像。未锁定的运行镜像会报错，避免静默漏包。
当前仅支持本发布清单中的 linux/amd64 镜像。

Helm 有自定义配置时，要在打包和安装时使用相同的 values：

```bash
python3 tools/package-offline.py --values values-single-node.yaml
```

工具不会把渲染后的 Secret 内容打印出来。自定义 values 会复制到离线包，
如果里面含凭证，请只向授权用户分发。不要将含凭证的 values 提交到公共仓库。
Compose 使用本目录的 .env 和 compose.yaml；修改配置后重新准备镜像包。

## 检查缓存或制作离线包

只检查本机 Docker 是否具备全部镜像且内容、架构匹配，不拉取：

```bash
python3 tools/package-offline.py --check
```

在有网络和 Docker 的准备机生成离线目录（目标目录必须不存在）：

```bash
python3 tools/package-offline.py --output ../lakehouse-offline
```

默认使用匹配的本地缓存，仅拉取缺失镜像。要验证全部镜像能够在线获取，使用
`--pull always`；仅从已有缓存打包，使用 `--pull never`。这些选项只在
`--output` 打包模式生效。缺镜像、内容不匹配和下载失败都会停止，并保留
INCOMPLETE 标记；失败目录不能用于安装，重试请选择新的输出目录。

```bash
python3 tools/package-offline.py --pull never --output ../lakehouse-offline
# Helm 自定义部署示例：
python3 tools/package-offline.py --values values-single-node.yaml --output ../helm-offline
```

产物包含 deployment/、images.tar.gz、offline-manifest.json、SHA256SUMS 和 INSTALL.txt。
打包后整个目录传到目标环境，先执行 `sha256sum -c SHA256SUMS`。

Compose：

```bash
gzip -dc images.tar.gz | docker load
cd deployment
docker compose up -d --pull never
docker compose logs -f observer-bootstrap
```

K3s：在每个实际运行组件的节点上导入（Docker 的缓存不会自动进入 K3s）：

```bash
gzip -dc images.tar.gz | sudo k3s ctr images import -
```

然后按照 INSTALL.txt 和 deployment/USAGE.txt，用本地 Chart 执行原生 Helm 安装。
自定义 values 已复制进 deployment/offline-inputs/，INSTALL.txt 会带上对应参数。
不使用 kubectl apply 代替 Helm，因此保留 Helm release、升级和卸载管理。

## 范围

- 本工具不安装 Docker、K3s、Helm、StorageClass 或 cert-manager；这些基础环境要另行准备，
  完全离线环境也需要准备它们的安装文件和镜像。
- 内置 MinIO 启用时自动包含其镜像和所需 mc；使用外部存储后按实际渲染结果收集。
- Helm 外部存储参数参见 USAGE.txt；本测试版 Compose 保留现有内置 MinIO 拓扑，
  没有增加一键切换外部存储或监控 profiles，不能仅删除镜像条目来停用服务。
- 本工具验证配置镜像的 Docker config ID 和架构；记录获取到的 RepoDigests。
  Docker config ID 与 registry manifest digest 是不同概念。
- 镜像归档使用一次 docker save，避免同一归档重复存储共享层。
- 许可证及声明见 THIRD-PARTY-NOTICES.md；部署配置包不等于第三方软件的重新授权。
