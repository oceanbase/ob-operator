# OceanSeek 共享存储与 LogService 回归测试

## 环境要求与构建

准备 Go 1.25、Node.js 18+ 和 Yarn 1。页面验收需要 Kubernetes、可用的 StorageClass、待测版本的 CRD 和 Operator，以及 OceanBase、LogService 和 S3 兼容对象存储。每次验收使用独立的命名空间，并为数据和日志分配独立的 Bucket 或不重叠的前缀。

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

页面验收使用同一版本的 Dashboard 前后端，以及 `charts/oceanbase-dashboard` 中的 Prometheus 采集配置和告警规则。独立 Dashboard Chart 通过 `image` 指定镜像，OceanSeek Chart 通过 `dashboard.image` 指定镜像。使用发布包时，按包内的 `image-lock.json` 核对实际部署的镜像。

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
