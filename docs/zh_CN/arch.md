# 架构设计

[English version](../en_US/arch.md) is available.

本文不涉及 OceanBase 数据库本身的架构和数据库管理说明，如需了解请参见[官网文档](https://www.oceanbase.com/docs/common-oceanbase-database-cn-1000000000217922)。

ob-operator 遵循 Kubernetes 的 [Operator 拓展范式](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/)，聚焦于自定义资源及其控制逻辑。ob-operator 使用 Operator 开发框架 [kubebuilder@v3](https://book.kubebuilder.io/introduction) 为基础进行开发，所以底层架构与 [kubebuilder 的架构](https://book.kubebuilder.io/architecture)相近。通过向 Kubernetes 控制平面全局注册 Controller Manager，下辖若干控制器和 Webhook，来对自定义的资源进行控制。

* 控制器通过监听特定资源的特定事件对事件做出响应，依据实现好的逻辑将资源的实际状态（Status）和期望状态（Spec）对齐；
* Webhook 主要有设定默认值和进行资源规约校验两部分功能，分别由 Defaulter 和 Validator 两个模块完成。资源规约校验过程主要防止出现 ob-operator 预期之外的资源被安装到集群当中，无法被正常调度。例如创建租户时如果指定的集群并不存在，则会在 apply 资源时就把错误抛出，而不是调度到一半才用事件或者日志的方式告知用户。

## 自定义资源

* OBCluster: OceanBase 集群资源
* OBZone: OceanBase Zone 资源，隶属于某个 OBCluster
* OBServer: OceanBase observer 资源，隶属于某个 OBZone
* OBParameter: 集群参数
* OBTenant: OceanBase 集群当中的租户，隶属于某个 OBCluster
* OBTenantBackupPolicy: 租户备份策略
* OBTenantBackup: 租户备份任务
* OBTenantRestore: 租户恢复任务
* OBTenantOperation: 租户运维操作

## 资源管理器

每种资源都由其对应的资源管理器进行统一调度，各个资源管理器都实现了接口ResourceManager，其定义如下。它包含了资源初始化，资源状态更新，资源任务获取，错误信息输出，资源删除操作等方面的方法。

```go
type ResourceManager interface {
  IsNewResource() bool
  IsDeleting() bool
  CheckAndUpdateFinalizers() error
  InitStatus()
  SetOperationContext(*v1alpha1.OperationContext)
  ClearTaskInfo()
  HandleFailure()
  FinishTask()
  UpdateStatus() error
  GetStatus() string
  GetTaskFunc(string) (func() error, error)
  GetTaskFlow() (*task.TaskFlow, error)
  PrintErrEvent(error)
  ArchiveResource()
}
```

ResourceManager 是典型的状态机模型，其调度资源的大致流程为：
1. 如果是新的资源，初始化其状态字段 `status.status`
2. 根据资源状态获取响应的任务流(`TaskFlow`)
  * 如果获取到的任务流非空，则将其存储在资源的 `status.operationContext` 当中，以较短的间隔轮询任务状态
    * 如果有任务待执行，则将任务提交到任务管理器，并将任务置于 `Pending` 状态并轮询任务状态
    * 如果任务成功，则执行下一个任务或者将资源置为下一个状态
    * 如果任务失败，则选择重试或者将资源置为错误状态
  * 如果获取到的任务流为空，表示当前资源正常运行没有变化，则以较长的间隔让该资源重新入队
3. 处理和响应资源的删除信息
4. 更新资源状态（`status.status` 和 `status.operationContext` 等字段都在此处更新）
5. 将调解结果返回给 ControllerManager，主要是返回重新入队间隔或者错误信息

## 任务流与全局任务管理器

Kubernetes 在内部采用控制循环和消息队列的方式来实现事件的采集和分发，事件由 Kubernetes 控制平面分发到各个 Controller Manager，Manager 再分发给各个 Controller 进行调解。控制器每次收到事件，会调用调解过程，为了避免可能的竞争冒险现象，控制器的调解 Worker 数量一般都为 `1`，也就是说只有前一个调解任务完成之后才哦能开启下一个调解任务，如果一个调解任务耗时太长则会阻塞住该类型资源其他事件的调解。所以一般来说 Kubernetes 的 Operator 模式或者说它的资源调度模式并不适合长调度任务。

为了解决这个问题，ob-operator 采用了任务流和全局任务管理器的方式来解决长调度问题。任务流由任务列表，当前执行的任务索引和任务状态信息组成；全局任务管理器则包含了两个 Map 结构：

* 工作集映射：`TaskID -> chan Result`，执行中或已结束但未读取结果的任务集合
* 结果缓存映射：`TaskID -> Result`，已结束（成功、失败）任务的结果集合

控制循环、资源管理器、任务管理器的关系如下图所示。

![ob-operator 资源调度过程](../img/ob-operator-arch.png)

## 资源与任务管理器的交互

任务流中的任务由资源管理器 `ResourceManager` 提交给全局的任务管理器 `TaskManager` 来执行，资源、资源管理器和任务管理器的大致关系和相互作用流程如下面的时序图所示：

<main>
  <pre class="mermaid">
sequenceDiagram
	participant r as Resource
	participant c as Controller (ResourceManager)
	participant t as TaskManager
	autonumber
	r->>c: Resource changes
	c->>t: Get task flow according to recourse status
	t->>t: Create goroutine to execute specific task
	t->>c: Return task ID to controller
	c->>r: Stores task ID and other task context in resource
	loop Watch task progress
		r->>c: Requeue and requeue
		c->>t: Checks the task status
		alt If task is still pending
			t->>c: Empty result
			c->>c: Continues loop and requeues resource with a shorter interval
		else If task is finished
			t->>c: Task results
			alt if no other tasks in flow
				c->>r: Updates status of resource
			else if there are other tasks in flow
				c->>r: Updates task context of resource
				c->>t: Watches progress of new task, back to [6] loop
			end
		end
	end
	t->>t: Clean maps
  </pre>
  <script type="module">
    import mermaid from 'https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.esm.min.mjs';
  </script>
</main>