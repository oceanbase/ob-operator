# 湖库 SS / LogService 测试交接

测试分支：`codex/ss-lakehouse-qa-20260921`。基线：`fd0f9362`（已有 LogService / oceanbase.ai CRD 和控制器）。此分支为功能验收版本，不代表生产发布或全仓库测试通过。

## 包含的改动

- Operator：LogService 日志 PVC namespace、异步任务快照与失败恢复修复；SS Zone 对象存储属性的 ADD/DROP 兼容处理。
- Dashboard：LogService 创建、列表、详情、已有 Zone 内副本扩缩容、引用中的删除保护。
- 湖库入口、四步创建向导、内嵌创建/关联 LogService、计算与缓存展示；OB 数据对象存储和 LS 日志对象存储分开配置。
- Endpoint / Bucket / Region / 前缀与高级 URL 双向切换、同命名空间 Secret 引用/创建、只读 HeadBucket 检查及独立权限控制。
- LS、对象存储、缓存专项采集、图表和告警；修复时间轴、无效事件时间、告警深链接、SS 向导初始白屏及静态资源缓存问题。
- 创建阶段要求 LS 初始节点总数为 3；SS 缓存盘至少 50 GiB，普通模式的 30 GiB 最小配置不变。均有前后端校验。

## 环境要求与构建

已验证组合为 OceanBase AI 4.6.2.0、LogService 1.3.0、S3 兼容 MinIO。镜像及其访问权限需由测试环境提供；本分支不包含数据库商业镜像、对象存储账号、真实 kubeconfig 或运行日志。

需要 Go 1.25、Node.js 18+ / Yarn 1，以及可运行对应镜像的 Kubernetes、可用 StorageClass、已安装的本分支 CRD / Operator。用独立命名空间、独立数据/日志 Bucket 或不重叠前缀做验收，不对已有业务集群做破坏性测试。

从仓库根目录构建 UI 与两个后端（二进制构建示例为 linux/amd64）：

```bash
cd ui
yarn install --frozen-lockfile
node --test tests/*.test.cjs
npm run build
cd ..
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/oceanbase-dashboard ./cmd/dashboard
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o bin/manager ./cmd/operator
```

需要构建镜像时使用已有的 `build/Dockerfile.dashboard` 和 `build/Dockerfile.operator`，分别打自己的测试标签，不覆盖正式镜像。以上快速二进制编译不注入发布版本元数据；正式构建流程见 `make/dashboard.mk` / `make/build.mk`。

部署时同时使用本分支 `charts/oceanbase-dashboard` 的 Prometheus 采集配置与告警规则，不能只更换 UI。现有 chart 的 Dashboard 镜像写在 `templates/bundle.yaml`，默认仍指向官方版本；测试部署必须在自己的部署清单/渲染流程里替换为新构建的镜像。不要误以为 checkout 分支或只改 values 就已部署新代码。使用独立登录凭据，不使用 chart 示例密码；本文不提供对现有环境直接执行的升级/删除命令。

## 可重复的定向测试

在仓库根目录的独立 shell 设置仅供单元测试的环境。下面的 kubeconfig 没有任何凭据，指向不可达地址，不能用于部署或真实 API 验收：

```bash
export KUBECONFIG="$PWD/tests/ss-test-kubeconfig.yaml"
export RBAC_POLICY_CONFIG_MAP=unit-test
export RBAC_POLICY_PATH=/dev/null

go test ./internal/dashboard/model/param \
  ./internal/dashboard/model/response \
  ./internal/dashboard/model/alarm/silence \
  ./internal/dashboard/business/logservice \
  ./internal/dashboard/business/objectstorage \
  ./internal/dashboard/business/ssmonitor \
  ./internal/dashboard/handler ./internal/dashboard/router

go test ./internal/dashboard/business/oceanbase \
  -run 'TestSharedStorage|TestGenerateSharedStorage'

go test ./internal/resource/oblogservicecluster \
  ./internal/resource/oblogservicezone \
  ./internal/resource/oblogservicenode \
  ./internal/resource/obzone ./pkg/oceanbase-sdk/operation
```

告警规则测试需要 Helm、Mike Farah yq v4 和 promtool。在仓库根目录渲染真实 chart 中的规则，再执行用例：

```bash
helm template dashboard-qa charts/oceanbase-dashboard |
  yq -r 'select(.kind == "ConfigMap" and .metadata.name == "dashboard-qa-prometheus-rules-conf") | .data["prometheus.rules"]' \
  > tests/ss-rules.generated.yaml
(cd tests && promtool test rules ss-alert-rules-test.yaml)
```

生成文件已忽略，不提交。测试覆盖 LS 副本不足告警触发/恢复、空闲缓存不误报、持续低命中率告警。

## 页面验收清单

| 场景 | 验收标准 |
| --- | --- |
| 正常登录、湖库入口、直接打开/刷新深链接 | 页面无白屏，无未处理 JavaScript 异常；告警规则不被重定向到事件页 |
| 四步向导、前进后退、普通/SS 切换 | 必填拦截，字段保留；普通模式不提交 SS 字段，SS 不提交本地 redoLog |
| 数据与日志对象存储 | Endpoint / Region 可复用；Bucket / 前缀独立填写，重叠有提示；高级 URL 往返一致 |
| 凭据与权限 | 仅能选择同命名空间 Secret；创建不覆盖已有 Secret；不回显密钥；无权限用户无写入口，API 同样拒绝 |
| HeadBucket 检查 | 成功/失败状态和检查时间可见，不泄露供应商原始错误或签名；没有对象写入/删除 |
| 创建 LogService | 1、2、4 初始节点被拒绝；3 节点可提交，最终 LS running、3 Ready LN、6 Bound PVC；未 running 不可关联 |
| 已有 Zone 内 LS 扩缩容 | 独立测试 LS 执行 3→4→3，期望副本、实际 LN、PVC 与页面一致；等待收敛后再做下一步 |
| 创建 SS 集群 | 30/49 GiB 缓存被拒绝，50 GiB 可创建；OB/Zone/Observer 最终 running，Pod Ready，核对资源与确认页一致 |
| LS 引用保护 | 被 OB 引用后禁止删除；不要用业务 LS 验证删除。取消向导不会自动删除已创建 LS/Secret |
| LS / SS 监控 | Ready/期望/Active LN、采集状态、对象存储计数/带宽可见；时间轴有效；无有效样本不能伪装为零 |
| 普通模式回归 | 原有创建、列表、详情、告警正常，最小 data 仍为 30 GiB |

每个写操作均记录页面结果、HTTP 状态、CR 状态及 Pods/PVC 终态。不能只以请求 200 或 Pod Ready 代替完整业务就绪。测试完成后，仅按明确的命名空间和资源名清理自建资源，保留所需证据；不要强制删除 finalizer。

## 已验收范围和已知限制

2026-09-21 已完成真实登录后的向导、LS 创建和 3→4→3、SS 50 GiB 创建至 running、关联删除保护及专项图表验收；前端回归 25/25、构建通过，定向 Go 测试通过。现场最终镜像发布检查 33/33、补充状态检查 5/5 通过。

边界需保留：最后一次页面创建使用新 UI 和上一版后端；最终后端增加的缓存下限由 Go 测试覆盖。最终整包重启后未再次完成登录后的全流程验收，测试同学仍应按上表在本分支独立构建部署后复测。

- LS 仅支持已有 Zone 内副本扩缩容；不开放在线加减 Zone、换镜像或调整 CPU/内存/磁盘。
- SS 数据存储位置、LS 关联为创建时配置，不支持在线迁移 Bucket、切换 LS、凭据轮换或 SS 镜像升级。
- 当前表单支持已验证 S3 子集；不声称支持所有 OSS 供应商、STS 或所有数据库版本。
- HeadBucket 只验证 Dashboard 后端对此 Bucket 的本次访问，不证明对象读写/删除、前缀权限或所有数据库节点的连通性。
- 压测、节点故障注入、多供应商对象存储兼容和完整业务 SQL 读写不在本轮页面验收范围内。
- 全量 TypeScript 检查仍有 486 项既有错误；全量 OceanBase business Ginkgo 有两项既有失败（affinity operator 空值/In、usage 13/6）。不将定向测试等同于全仓库测试通过。
- Dashboard Pod 重启可能使现有登录失效，需要正常重新登录，不应重置账号来绕过。

现场日志、截图、一次性主机脚本与已发布二进制仅在测试环境保留，不随分支提交；单元测试中的地址和密钥均为示例或哨兵值。
