# 变更日志

## 2.3.0 (发布于 2024.10.14)

### 新增特性

1. 支持跨 K8s 集群调度 OceanBase 集群
2. 支持设置腾讯云 COS、AWS s3 以及 s3 兼容的对象存储服务作为数据备份的介质
3. 支持删除特定的 OBServer
4. 支持根据场景设置优化 OceanBase 集群的系统参数和变量
5. 支持将大部分 K8s 内置的 Pod 字段设置到 OBServer 中

### 缺陷修复

1. 修复 2-2-2 集群滚动替换 OBServer 时可能出现卡住的问题

### 功能优化

1. 补充了几种新的 `OBTenantOperation` 类型用于执行常见操作，如创建或删除资源池、设置 Unit Number 等

## 2.2.2 (发布于 2024.06.18)

### 新增特性

1. 支持创建集群时给 `storageClass` 字段留空以使用默认存储类
2. 支持通过 `/pause-reconciling: "true"` 注解暂停 CRD 的调度，可能解决某些问题
3. 支持在运行中的 OBClusters 中更改存储类
4. 新增 CRD `OBClusterOperation` 用于对 OBClusters 进行常见运维操作

### 缺陷修复

1. 修复集群减半缩容后可能出现的失活问题（例如从 2-2-2 立即缩容为 1-1-1）

## 2.2.1 (发布于 2024.05.15)

### 新增特性

1. 支持通过环境变量细粒度地配置 ob-operator
2. 能够保持 Pod IP 的情况下支持集群启动后挂载备份卷
3. 支持通过增加特定注解来增加资源的删除保护
4. 支持迁移 K8s 外的 OceanBase 集群至 K8s 中让 ob-operator 接管

### 缺陷修复

1. 修复无法创建跨命名空间和跨集群的备租户的问题
2. 修复当 OceanBase 集群不可访问时恢复 OBServer 的问题

### 优化

1. 部署集群前检查 Clog 存储卷兼容性
2. 优化服务模式下创建集群时的版本检查

## 2.2.0 (发布于 2024.03.28)

### 新增特性

1. 支持通过创建 `service` 模式的集群，在不依赖特定的 CNI 插件的情况下保持 OBServer 的通信 IP 地址不变
2. 支持使用 CRD `OBResourceRescue` 提供的`删除`、`重置`、`跳过`和`重试`操作矫正其他相关 CRD 资源的状态
3. 支持在集群运行过程中动态扩容 PVC

### 缺陷修复

1. 修复集群剩余资源不足情况下创建资源超限（CPU 和内存等）的租户时不断报错重试的问题

### 功能优化

1. 加速子资源发生变化时的资源调解过程
2. 通过 Prometheus 标准接口暴露 Controller Manager 的监控指标

## 2.1.2 版本 (发布于 2024.01.24)

### 新增特性

1. 支持使用 `oceanbase.oceanbase.com/mode`: `standalone` 注解创建 standalone 集群
2. 支持 standalone 模式集群的垂直扩展，调整 CPU 和 Memory 资源
3. 支持使用 `oceanbase.oceanbase.com/single-pvc`: `true` 注解创建使用单一 PVC 挂载的节点
4. 支持通过 `spec.serviceAccount` 字段向集群的 Pod 绑定指定的 ServiceAccount

### 缺陷修复

1. 如果 ob-operator 的 Pod 重启，正在恢复的 OBServer Pod 无法正常恢复
2. 当 OBZone 调整副本数量时，即使已经有正在删除的 OBServer，OBZone 仍会随机删除它的节点

### 功能优化

1. 优化任务管理器，增加令牌池限制最大运行中任务的数量
2. 优化数据库连接池，使用带过期时间的 LRU 缓存数据库连接
3. 精简集群初始化必要的用户凭证为 root，自动创建其他用户凭证 (proxyro、monitor 和 operator)
4. 精简租户所需的用户凭证，root 和 standbyro 均为选填，standbyro 若不传递，会默认创建
5. 在创建集群的 Pod 之前等待所需镜像拉取到本地

## 2.1.1 版本 (发布于 2023.12.20)

### 新增特性

1. 支持向 `OBCluster` 资源添加 `oceanbase.oceanbase.com/independent-pvc-lifecycle` 注解使得 `OBCluster` 被删除时 PVC 得以保留
2. `OBTenantOperation` 资源支持 `Upgrade` 操作
3. 新建 `OBCluster` 资源时所携带的 parameters 参数将作为启动参数 `optstr` 传递给 observer 进程

### 缺陷修复

1. 显式设置 `memory_limit` 参数，修复某些容器运行时（CRI）中无法限制内存资源的问题
2. 初始化集群时传递初始化参数，避免在集群初始化后再设置参数所需的长时间等待

### 功能优化

1. 减少初始的 `datafile_size` 参数值，使用步进的方式按需增大数据文件磁盘用量
2. 加强资源规格校验，尤其针对 `OBCluster` 和 `OBTenant`

## 2.1.0 版本（发布于 2023.11.20）

### 新增特性

1. 集群配置中新增亲和性和容忍性的选项
2. 支持从备份数据恢复出租户和创建备租户
3. 提供 ARM 架构镜像

### 缺陷修复

1. 修复 map 并发写可能出现的问题
2. 修复租户备份过程中使用过期的数据库连接的问题
3. 修复 OceanBase 4.2.1 SQL 语法兼容性问题

### 功能优化

1. 采用回退机制优化出错任务重试过程
2. 增加事件打印，优化日志输出

## 2.0.0 版本（发布于 2023.09.26）

### 新增特性

1. 支持 OceanBase 集群的管理功能
2. 支持 OceanBase 租户的管理功能
3. 使用 obagent 来监控 OceanBase
