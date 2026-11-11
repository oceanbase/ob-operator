---
name: ob-operator-dima
description: 处理 ob-operator 的 DIMA 工单，完成复现、修复、三个 subagent 独立测试和审查、按授权提交及回填，测试通过后标记 Fixed。用于用户要求按 DIMA 流程处理工作项，或在当前项目提供待处理的 DIMA 链接或 ID；仅分析等范围限制优先，流程配置和状态询问不启动修复。
---

# ob-operator DIMA 修复流水线

参考 OBD 的 DIMA 流程，适配本仓库的 Go/Kubernetes 实现、仓库内测试和实际 Git 托管平台。

## 执行范围

- 用户要求“按 DIMA 流水线处理”时，完整执行读取工单及关联文档、RED、修复、GREEN、三个 subagent 独立测试和审查、按授权提交、DIMA 回填及 Fixed 状态确认。会话中已有的授权持续有效，不重复确认已授权步骤。
- 用户已明确设定长期规则：以后提供待修复 DIMA 工单，修复完成必须启动 3 个 subagent 实际测试；三方测试通过、修复完成并核对提交后，将该工单标记为 Fixed 并读回确认，无需再次询问。三个测试 agent 可同时承担下述三方审查，不要求另起重复的一组三人。
- 用户只给出待处理的工单链接或 ID 时，先读取工作项确定问题和范围；按用户的处理意图继续。链接作为背景、状态查询、仅分析或修改本流程不授权提交代码或外部评论。
- “不提交”时仍完成允许的代码、测试和审查，跳过 commit、push、PR/MR 和依赖提交链接的回填。“不回填”仅跳过 DIMA 写入。明确限制优先于默认流程。
- 未明确的目标分支、测试环境或访问权限先从工作项、当前会话和仓库取证；只有无法确定且影响下一步时询问，并继续其他不依赖该信息的工作。
- 用户指定直接推送测试分支、不提 PR/MR 时，沿用已确认的目标分支，跳过 PR/MR 及依赖其 URL 的回填；仍执行三个 subagent 测试和 Fixed 状态确认。
- 该流程不包含自动合并、发布版本或部署生产环境；工单状态只按上述授权改为 Fixed，不自动 Closed。

## 1. 读取工单和确认目标

1. 从 URL 的 `openWorkItemId` 提取 ID，保留可访问的完整工单链接。裸 ID 需从工具返回或 DIMA 页面取得真实链接，不凭空拼接项目/迭代 ID。
2. 执行 `dima workitem get <id> --with-relations -o json`，按需读评论，记录所属项目、受影响组件、版本、迭代、验收条件和已有 PR/MR。
3. 描述含语雀链接时，重新读取最新正文；方法见 [DIMA 和语雀接入](references/dima.md)。不能以标题或旧会话摘要替代正文。
4. 检查 `git status --short`、`git remote -v`、远端分支和已有 PR/MR。当前仓库使用 GitHub，但每次以实际 remote 为准；不要套用 GitLab push options。
5. 目标分支按用户指定、工单组件/版本、实际远端分支共同确定：Operator 常见 `X.Y.Z_release`，Dashboard 常见 `ds_X.Y.Z_release`，CLI 常见 `cli_X.Y.Z_release`。`master` 仅用于确认以主线为目标的工作。迭代版本未必等于组件版本；无唯一匹配时先明确，禁止默认 `master` 或虚构 `dev_X.Y.Z`。
6. 从确认的目标分支创建 `codex/dima-<id>-<short-description>` 修复分支。需要隔离时使用 worktree，保留用户未提交内容。记录基线 SHA；已有修复分支/PR 则继续原有工作，避免重复创建。

## 2. 定位根因并建立 RED

先读相关路径和已有测试，确定最小复现，再修改业务代码。

| 问题范围 | 主要位置 | 回归方式 |
| --- | --- | --- |
| CRD、Webhook 校验 | `api/`, `config/crd/`, `config/webhook/` | 包内 Go/Ginkgo + envtest |
| Reconcile、调度、状态机 | `internal/controller/`, `internal/resource/`, `pkg/task/` | 包内用例；涉及真实生命周期时补 K8s 场景 |
| SDK、公共逻辑、配置 | `pkg/`, `internal/config/` | 受影响包的 Go 用例；检查外部服务依赖 |
| Dashboard、CLI | `internal/dashboard/`, `internal/cli/`, `ui/` | 受影响包/界面行为与对应构建检查 |
| 集群、租户、备份恢复等 | `tests/case_p*/`, `tests/config/` | 对应 Shell 场景及资源/数据库状态断言 |

- Go 用例与代码同仓、就近放置；场景用例放 `tests/`。默认一个 PR/MR 同时提交修复与回归用例，不沿用 OBD 的外置 case 仓库和分支。
- 在未修复基线上运行新增用例，保存命令、基线 SHA、环境/镜像、退出码和失败断言。缺依赖、网络失败、编译失败或环境未就绪不能作为业务 bug 的 RED。
- 使用 `go test -count=1` 避免复用缓存；普通 Go 用例可用 `-run`，Ginkgo 场景需用 `-ginkgo.focus` 选择 spec。确认目标用例确实运行，零用例或 Skip 不算通过。
- 对 Go 行为修复提供针对性自动回归；文档等无法构造运行时 RED 的变更说明适用性，使用对应校验，不制造虚假 FAIL。

## 3. 最小修复并验证 GREEN

只修根因相关代码。使用相同测试、配置和断言复跑，记录修复后 SHA 与 GREEN 证据；不得降低断言来获得通过。

按目标分支的 Makefile/CI 准备工具。当前 CI 的基本入口为：

```bash
# 在独立工作目录中执行；Go 版本以 go.mod 为准，Go 安装工具需在 PATH 中。
make dashboard-bindata-gen dashboard-doc-gen
go mod vendor
make tools
make test
make build
```

- `make test` 只执行 `test-webhooks`、`test-package`、`test-internal`；额外运行受影响但未覆盖的包，不能将其称为全仓测试通过。
- Webhook/controller 的 envtest 依赖 `KUBEBUILDER_ASSETS`，版本从当前 Makefile 读取。额外运行 controller suite 时也要显式设置该变量；envtest 不等于启动真实 OceanBase。
- `internal/resource/resource_suite_test.go` 使用 `UseExistingCluster`；其他测试可能连接 Kubernetes、数据库或外部服务。先读 suite，再选择已有授权的测试环境，不盲跑 `go test ./...`。
- Operator 变更运行 `make build`；相关组件分别增加 `make dashboard`、`make okctl`、`make helper` 或 `make sql-analyzer`。SQL Analyzer 的 CGO/DuckDB 依赖按其构建文件处理。
- UI 变更按 `ui/package.json` 和 UI workflow 运行依赖安装、`yarn check-style`、`yarn lint`、`yarn build`，并验证受影响的用户行为。
- API/类型变更检查 `make manifests generate` 生成的 CRD、RBAC、DeepCopy 等差异。格式化/生成造成的无关变化不得混入提交。

### 真实 Kubernetes 场景

1. 使用已确认的测试 kubeconfig/context、隔离 namespace 和存储资源；记录 Operator 镜像/提交及 OceanBase 版本。不要使用 OBD 的固定主机或 Pod 地址。
2. 先检查场景脚本和依赖配置，再在 `tests/` 内运行。每个阶段使用对应的基线/修复 Operator 镜像，确认 rollout 和运行中的镜像标识后测试。
3. `tests/env.sh` 及场景生成的 `env_vars.sh` 可能含凭据；使用本地测试配置，不提交、不输出到 PR/DIMA。
4. `tests/run.sh` 不可靠地聚合失败，部分 case 只打印 `case failed`。必须逐个检查实际断言、日志、CR 状态和必要的 SQL 结果。新增/修改的场景应让断言失败返回非零，不能只检查外层退出码。
5. 有些场景依赖前序 case 的资源和 `env_vars.sh`。确认依赖后按序运行；独立上下文/目录/namespace 之外不并发同组场景。清理仅限此次创建且不再被后续 case 使用的资源。

按变更风险完成相关模块回归及所需集成场景。无法运行的检查记录为未验证和具体原因，继续可运行部分，不标记 GREEN 或测试门禁通过。

## 4. 三个 subagent 独立测试和审查

代码和测试稳定后，以同一最新提交 SHA 和相对目标分支的完整 diff，启动三个独立 subagent：

1. **功能/控制器测试与审查**：实际运行缺陷复现和修复回归，核对根因、reconcile 幂等、重试、状态转换、finalizer、资源生命周期和版本兼容。
2. **Go/接口测试与审查**：实际运行相关接口/边界/兼容用例，按风险补充 race、构建等检查；覆盖错误处理、context、nil、Kubernetes API/Webhook/CRD 契约；前端问题相应覆盖 TypeScript 和用户交互路径。
3. **安全/回归测试与审查**：实际运行相关回归及适用的权限、输入边界用例，核对 namespace 隔离、Secret/RBAC、RED/GREEN 证据和断言有效性。

根据缺陷调整三人的测试分工，但每人必须实际执行适用测试或可执行验证，不能仅读代码、引用主 agent 的日志或给出审查结论。每人记录同一修复 SHA、执行命令、实际用例/断言、退出码、日志位置、发现及明确结论。主 agent 汇总和修复，不代替任何一人的测试或审查。避免并发生成同一产物或覆盖测试状态。

- 存在可操作问题、测试失败或结论分歧时核实并修复，对最新提交重新取得三方测试和审查结论；不能多数票通过。
- 三方实际测试全部通过且均无未解决 findings 才通过。扩大回归发现既有失败时，必须在未修复基线独立重现并记录相同失败，确认与本次修复无关；这类失败不能宣称通过，也不替代本次必需测试。缺陷相关必需测试未运行、被跳过或失败时，不得改为 Fixed。
- 并发槽不足时顺序启动三个独立 subagent；工具不可用时标记测试/审查未完成，不能让主 agent 冒充三个 reviewer，也不能改为 Fixed。
- 同一阻塞连续三次且无新证据时停止重复尝试，保留完成状态与待解决项；不能为结束循环把阻塞写成 PASS。

## 5. 提交 PR/MR 并审查远端 HEAD

在用户授权的提交范围内：

- 用户要求直接推送、不提 PR/MR 时，只提交并推送到已确认分支，读回远端 SHA；同一 SHA 的三方测试和审查可复用。远端提交发生变化且包含新修复时，重新验证。跳过本节其余 PR/MR 专属步骤，不因无 PR 阻塞 Fixed。

- 只 stage 本任务文件，提交前检查暂存 diff，不夹带用户文件、凭据或测试产物。代码和 case 默认在同一个 PR/MR。
- 向已经确认的 remote/fork 推送，显式指定目标仓库、目标分支、head 分支。GitHub 使用可用的连接器、`gh` 或浏览器；GitLab remote 才使用对应 MR 工具。工具缺失不能当成跳过此阶段。
- 创建前按 head 分支查询已有 PR/MR；已有则更新。描述包含完整 DIMA 链接、问题与修复行为、RED/GREEN、回归证据和尚未验证项。多行正文用工具结构化参数或 `--body-file`，不要用 shell 拼接未经转义内容。
- 当前 upstream 为公共 GitHub 仓库。公开描述只包含适合公开的技术说明及已授权工单引用，不复制内部语雀正文、环境地址或原始敏感日志；私有内容留在授权的内部位置。
- 创建后回读远端 `base`、`head`、SHA、diff 和 checks，对实际最新 HEAD 再做一次上述三方独立审查。发现问题时更新同一 PR/MR，重新测试及全体审查，直到一致。
- 检查配置要求的 CI checks；pending/失败不可写成成功，无匹配 workflow 时明确记录，不能凭空声称 CI 通过。修复导致 HEAD 改变后，旧 HEAD 的 review/checks 不适用。
- 在 Codex 中创建 PR 后调用可用的 `attach_artifact` 附加链接。

## 6. DIMA 回填和验收

用户授权完整流程或回填且三方审查通过后，按 [回填步骤](references/dima.md) 执行并独立回读：

1. PR/MR 描述含准确、完整的 DIMA 链接。
2. DIMA 评论含准确的修复 PR/MR 链接；先查重，复用已存在的匹配评论。
3. DIMA `relations` 中存在 `relationType == "URL"` 且 `url` 精确匹配 PR/MR 的正式关联。

GitHub 不一定具备 OBD GitLab 的自动关联集成。先完成评论，再检查正式链接；必要时用 DIMA 页面添加，不能假设集成会生成关联，也不能用评论替代正式关联。`dima workitem link add` 只用于工作项关系，不能拿来添加 PR URL。

同一外部操作返回不确定结果时，先回读再决定重试；权限/登录失败或最多三次读回仍未确认时，记录未完成项和恢复入口，不无限重试、不重复创建 PR 或评论。

### 测试通过后标记 Fixed

确认三个 subagent 均对最终修复提交实际测试通过、无未解决发现，且授权范围内的提交已完成并读回。随后按 [状态更新步骤](references/dima.md#测试通过后更新-fixed) 使用 DIMA CLI 查询该工单可用状态，更新为 Fixed 并独立读回。不要只在聊天中宣称已修复，也不要等待用户再次提醒。明确不回填、只分析等限制优先。

最终逐项报告：目标分支、基线/修复 SHA、RED、GREEN、相关回归、三个 subagent 的实跑测试和审查、PR/MR 与 CI、描述链接、DIMA 评论 ID、正式 URL 关联、Fixed 状态读回。只有所需项全部有通过证据，完整流水线才完成；用户排除项标记“按用户要求跳过”，环境/权限阻塞项标记“未完成”。
