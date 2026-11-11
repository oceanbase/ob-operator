# 用 Qoder 部署 OceanBase：AI 驱动的数据库部署新范式

> 从一句话指令到生产就绪的数据库集群，Qoder 让 OceanBase 部署变得前所未有的简单。

## 为什么是 Qoder + OceanBase？

**OceanBase** 是蚂蚁集团自主研发的原生分布式关系数据库，具备金融级高可用、透明水平扩展、高性能低成本等核心特性，已连续多年刷新 TPC-C 和 TPC-H 世界纪录。OceanBase 社区版（OceanBase CE）完全开源，是企业级分布式数据库的首选方案。

**Qoder** 是一款智能 Agent 编程平台，专为真实软件开发而设计。通过 MCP（Model Context Protocol）协议和内置的 OceanBase 技能库，Qoder 能够以自然语言交互的方式，自动完成 OceanBase 的规划、部署、配置和验证全流程——无需手动翻阅文档、无需逐条敲入命令。

### 传统部署 vs Qoder 部署

| 对比维度 | 传统手动部署 | Qoder AI 部署 |
|---------|------------|--------------|
| **环境检查** | 手动登录服务器，逐条执行命令检查 OS、CPU、内存、磁盘 | 一句话指令，自动 SSH 并完成全量环境巡检 |
| **版本选择** | 手动查询镜像源，比对版本兼容性 | 自动识别 OS 架构（el7/el8），选择最优兼容版本 |
| **配置编写** | 查阅文档，手写 YAML 配置文件 | 根据服务器资源自动生成最优配置 |
| **部署执行** | 逐条执行 obd 命令，人工确认每一步 | 自动执行部署全流程，实时反馈进度 |
| **故障排查** | 翻阅日志，搜索引擎查错 | 自动分析日志，定位问题根因并修复 |
| **验证连通** | 手动执行 SQL 测试 | 自动完成建库、建表、增删改查全链路验证 |
| **开机自启** | 手动编写 systemd 服务 | 自动创建并启用 systemd 服务 |
| **总耗时** | 30-60 分钟（含查文档） | 5-10 分钟 |

## 两种部署路径

Qoder 支持 OceanBase 的两种主流部署方式，用户可根据实际场景选择：

### 路径一：Kubernetes 部署（ob-operator）

适用于已拥有 Kubernetes 集群的环境，通过 **ob-operator** 以云原生方式管理 OceanBase 集群全生命周期。

**ob-operator** 是满足 Kubernetes Operator 扩展范式的自动化工具，提供以下核心能力：

- **集群管理**：集群自举、拓扑调整、扩缩容、升级、参数调优
- **租户管理**：创建租户、调整拓扑、资源单元管理、密码重置
- **备份恢复**：周期性备份至 OSS/NFS、数据恢复
- **物理备库**：备租户创建/激活、主备切换
- **故障恢复**：单节点及集群级故障恢复
- **可视化仪表盘**：Web UI 管理和监控集群拓扑

#### 快速部署流程

只需对 Qoder 说：**"帮我在 K8s 集群上部署 OceanBase"**，Qoder 将自动执行：

```bash
# 1. 部署 cert-manager（ob-operator 依赖）
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/stable/deploy/cert-manager.yaml

# 2. 部署 ob-operator
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/stable/deploy/operator.yaml

# 3. 创建密码 Secret
kubectl create secret generic root-password --from-literal=password='root_password'

# 4. 部署 OceanBase 集群
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/stable/example/quickstart/obcluster.yaml

# 5. 验证集群状态
kubectl get obclusters.oceanbase.oceanbase.com test
```

集群配置示例（Qoder 可根据实际资源自动生成）：

```yaml
apiVersion: oceanbase.oceanbase.com/v1alpha1
kind: OBCluster
metadata:
  name: test
  namespace: default
spec:
  clusterName: obcluster
  clusterId: 1
  userSecrets:
    root: root-password
  topology:
    - zone: zone1
      replica: 1
  observer:
    image: oceanbase/oceanbase-cloud-native:4.2.1.6-106000012024042515
    resource:
      cpu: 2
      memory: 10Gi
    storage:
      dataStorage:
        storageClass: local-path
        size: 50Gi
      redoLogStorage:
        storageClass: local-path
        size: 50Gi
      logStorage:
        storageClass: local-path
        size: 20Gi
  parameters:
  - name: system_memory
    value: 1G
  - name: "__min_full_resource_pool_memory"
    value: "2147483648"
```

#### 存储兼容性

ob-operator 已验证以下存储方案的兼容性：

| 存储方案 | 兼容性 | 适用场景 |
|---------|-------|---------|
| local-path-provisioner | ✅ | 开发测试环境 |
| Rook RBD (Block) | ✅ | 生产环境 |
| Longhorn | ✅ | 生产环境 |
| JuiceFS | ✅ | 生产环境 |
| OpenEBS (cStor) | ✅ | 生产环境 |
| NFS | ❌ | 不推荐（无法回收租户资源） |

### 路径二：裸金属部署（obd）

适用于没有 Kubernetes 的裸金属服务器或虚拟机环境，通过 **obd（OceanBase Deployer）** 进行部署。

#### 实战案例：一句话部署 OceanBase CE 4.5.0.0

对 Qoder 说：**"帮我在 root@11.124.9.67 部署一下 oceanbase"**

Qoder 自动完成的全流程：

**第一步：环境巡检**

Qoder 自动 SSH 登录目标服务器，完成以下检查：
- 操作系统识别（AliOS 7.2 / CentOS 7 / Ubuntu 等）
- CPU 架构与核心数（x86_64, 8 核）
- 内存总量与可用量（62G 总内存）
- 磁盘空间与挂载点（系统盘 60G + 数据盘 1.8T）
- 网络接口与 IP 地址
- 端口占用情况（2881/2882/2886 是否空闲）
- 已安装软件检测（obd 版本、现有集群状态）

**第二步：智能版本选择**

Qoder 根据 OS 版本自动选择兼容的 RPM 包：
- AliOS 7.2 → GLIBC 2.17 → 选择 **el7** 包
- 自动从镜像源选取最新兼容版本：**oceanbase-ce 4.5.0.0**

**第三步：自动配置生成**

根据服务器实际资源，Qoder 自动生成最优配置：

```yaml
oceanbase-ce:
  servers:
  - name: server1
    ip: 11.124.9.67
  global:
    devname: eth0
    mysql_port: 2881        # MySQL 连接端口
    rpc_port: 2882          # 内部 RPC 端口
    obshell_port: 2886      # Web 管理控制台
    home_path: /data/1/oceanbase-ce
    data_dir: /data/1/oceanbase-ce/store    # 数据目录（1.8T 数据盘）
    redo_dir: /data/1/oceanbase-ce/store
    memory_limit: 30G       # 内存限制（总 62G，预留给 K8s/Docker）
    system_memory: 2G       # 系统保留内存
    cpu_count: 6            # CPU 核心数（总 8 核）
    datafile_size: 20G      # 初始数据文件大小
    datafile_maxsize: 200G  # 数据文件最大值
    log_disk_size: 50G      # 日志磁盘大小
    production_mode: false
    zone: zone1
    cluster_id: 1
    root_password: root123
    appname: ob_cluster
```

**第四步：自动部署与启动**

```bash
# 部署集群
obd cluster deploy ob_cluster -c /root/ob_deploy.yaml -f

# 启动集群
obd cluster start ob_cluster
```

Qoder 实时监控部署进度，自动处理启动过程中的告警和异常。

**第五步：全链路验证**

Qoder 自动执行验证：
- 数据库版本查询：`SELECT VERSION();` → `5.7.25-OceanBase_CE-v4.5.0.0`
- 集群状态检查：`SELECT zone, svr_ip, status FROM DBA_OB_SERVERS;` → `ACTIVE`
- 功能测试：建库 → 建表 → 插入 → 查询 → 清理
- Web 控制台验证：obshell dashboard HTTP 200

**第六步：开机自启配置**

Qoder 自动创建 systemd 服务，确保服务器重启后数据库自动恢复：

```ini
[Unit]
Description=OceanBase Cluster (obd managed)
After=network.target

[Service]
Type=oneshot
ExecStart=/usr/bin/obd cluster start ob_cluster
ExecStop=/usr/bin/obd cluster stop ob_cluster
RemainAfterExit=yes
User=root

[Install]
WantedBy=multi-user.target
```

#### 部署结果

| 项目 | 值 |
|------|-----|
| OceanBase 版本 | 4.5.0.0 (CE) |
| 集群名称 | ob_cluster |
| MySQL 端口 | 2881 |
| RPC 端口 | 2882 |
| Web 控制台 | http://11.124.9.67:2886 |
| root 密码 | root123 |
| 数据目录 | /data/1/oceanbase-ce（1.8T 数据盘） |
| 内存限制 | 30G |
| CPU | 6 核 |
| 状态 | ACTIVE |
| 开机自启 | 已启用（systemd） |

## Qoder 部署 OceanBase 的核心优势

### 1. 自然语言驱动

无需记忆复杂的命令行参数和配置文件格式。只需用自然语言描述需求：

> "帮我在 root@192.168.1.100 部署一个 OceanBase 集群，内存限制 16G，3 个 zone"

Qoder 自动完成从环境检查到集群启动的全流程。

### 2. 环境感知与智能决策

Qoder 能够自动识别：
- **操作系统版本**：选择 el7/el8 对应的 RPM 包
- **GLIBC 版本**：避免版本不兼容问题
- **网络环境**：自动识别网卡名称和 IP 地址
- **资源状况**：根据可用内存和 CPU 智能分配资源
- **端口冲突**：自动检测端口占用并选择可用端口
- **已有部署**：避免与现有集群冲突

### 3. 全自动故障恢复

部署过程中遇到问题时，Qoder 能够：
- 自动分析 obd 和 observer 日志
- 识别错误码并定位根因
- 自动调整配置并重试
- 在版本不兼容时自动降级方案（如移除不兼容的 obproxy 组件）

### 4. MCP 协议深度集成

Qoder 通过 MCP（Model Context Protocol）与 OceanBase 生态深度集成：
- **oceanbase-deploy skill**：集群部署与生命周期管理
- **cluster-management skill**：集群启停、升级、扩缩容
- **tenant-management skill**：租户创建、备份恢复
- **testing-and-benchmark skill**：性能测试（Sysbench/TPC-C/TPC-H）

### 5. 部署后持续运维

Qoder 不仅仅是一个部署工具，更是持续的运维助手：
- 创建租户和用户
- 配置备份策略
- 执行性能基准测试
- 监控集群健康状态
- 故障排查与日志分析

## OceanBase 部署后的常用操作

通过 Qoder，以下操作同样只需一句话：

| 操作 | 自然语言指令 |
|------|------------|
| 创建租户 | "创建一个名为 test_tenant 的租户，分配 2C 4G 资源" |
| 创建用户 | "在 test_tenant 下创建用户 appuser，并安全设置密码" |
| 配置备份 | "配置每天凌晨 2 点备份到 NFS 目录 /backup/ob" |
| 性能测试 | "对 OceanBase 跑一轮 Sysbench oltp_read_write 测试" |
| 查看状态 | "查看 OceanBase 集群状态和所有租户信息" |
| 停止集群 | "停止 OceanBase 集群" |
| 升级集群 | "将 OceanBase 升级到 4.5.0.0" |

## 适用场景

| 场景 | 推荐部署路径 | 说明 |
|------|------------|------|
| **开发测试** | K8s + ob-operator | 快速拉起、弹性伸缩、一键销毁 |
| **生产环境（云原生）** | K8s + ob-operator | 自愈、滚动升级、存储兼容性强 |
| **生产环境（裸金属）** | obd 裸金属部署 | 性能最优、资源独占、运维简单 |
| **混合云** | 两种结合 | K8s 管理应用层，裸金属运行数据库层 |
| **个人学习** | obd 单节点部署 | 资源需求低（2C 10G 即可） |

## 环境要求

### K8s 部署（ob-operator）

- Kubernetes 集群 v1.18+
- kubectl v1.18+
- cert-manager（ob-operator 依赖）
- 至少 2 CPU、10G 内存、100G 存储
- 兼容的存储方案（推荐 local-path / Rook RBD / Longhorn / JuiceFS）

### 裸金属部署（obd）

- Linux 服务器（CentOS 7+ / Ubuntu 18.04+ / AliOS 等）
- 至少 2 CPU、8G 内存、50G 存储
- root 或 sudo 权限
- 网络互通

## 开始使用

### 方式一：在 Qoder 中直接部署

打开 Qoder，输入：

> "帮我在 [服务器地址] 部署 OceanBase"

Qoder 将自动完成全流程部署。

### 方式二：结合 ob-operator 项目开发

在 Qoder 中打开 ob-operator 项目仓库，即可：
- 阅读和修改 Operator 源码
- 调试控制器逻辑
- 构建和发布自定义镜像
- 编写和运行单元测试

### 方式三：通过 Helm Chart 部署

```bash
helm repo add ob-operator https://oceanbase.github.io/ob-operator/
helm repo update
helm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace
helm install oceanbase-dashboard ob-operator/oceanbase-dashboard
```

## 结语

Qoder 与 OceanBase 的结合，代表了 AI 时代数据库运维的新范式。从环境检查、版本选择、配置生成、部署执行到故障排查，Qoder 以自然语言为接口，以 MCP 协议为桥梁，将原本需要 30-60 分钟的手动部署流程压缩到 5-10 分钟。

无论是 Kubernetes 云原生场景下的 ob-operator 部署，还是裸金属服务器上的 obd 部署，Qoder 都能根据实际环境智能选择最优方案，让开发者专注于业务逻辑而非基础设施运维。

**立即体验**：打开 Qoder，输入 "帮我部署 OceanBase"，开启 AI 驱动的数据库部署之旅。

---

*本文基于 ob-operator 开源项目（https://github.com/oceanbase/ob-operator）和 OceanBase 社区版编写。ob-operator 使用 MulanPSL-2.0 许可证。*
