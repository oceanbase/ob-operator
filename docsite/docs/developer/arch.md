# Architecture

This document does not cover the architecture and instructions for managing the OceanBase database itself. If you want to learn more, please refer to the [official documentation](https://www.oceanbase.com/docs/common-oceanbase-database-cn-1000000000217922).

ob-operator follows the [Operator pattern of Kubernetes](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/), focusing on custom resources and their control logic. It is developed based on the Kubernetes Operator development framework, [kubebuilder@v3](https://book.kubebuilder.io/introduction), making its underlying architecture similar to [that of kubebuilder](https://book.kubebuilder.io/architecture). By globally registering a Controller Manager from the Kubernetes control plane and overseeing multiple controllers and webhooks, ob-operator controls custom resources like OBCluster and OBTenant etc.

* Controllers respond to specific events of specific resources and align the actual state (Status) with the desired state (Spec) based on implemented logic.
* Webhooks serve two functions: setting default values and performing resource validation. These tasks are accomplished by the Defaulter and Validator, respectively. Resource validation prevents unexpected resources from being installed in the cluster, ensuring proper scheduling by ob-operator. For example, if a specified cluster does not exist when creating a tenant, an error is thrown when applying the resource, rather than informing the user through events or logs halfway through the scheduling process.

## Custom Resources

* OBCluster: Represents an OceanBase cluster.
* OBZone: Represents an OceanBase zone that belongs to an OBCluster.
* OBServer: Represents an OceanBase observer resource that belongs to an OBZone.
* OBParameter: Represents cluster and tenant parameters.
* OBTenant: Represents a tenant in the OceanBase cluster, which belongs to an OBCluster.
* OBTenantBackupPolicy: Represents a scheduling backup policy for a tenant.
* OBTenantBackup: Represents a backup task for a tenant.
* OBTenantRestore: Represents a restore task for a tenant.
* OBTenantOperation: Represents operational tasks for a tenant.
* OBTenantVariable: Represents tenant variables.
* OBClusterOperation: Represents operational tasks for a cluster.
* OBResourceRescue: Represents a resource rescue task, which is used to recover resources that are stuck in an error state.
* K8sCluster: Represents a Kubernetes cluster that is managed by ob-operator.

## Diagram

The following diagram illustrates layers of ob-operator and the relationship among the custom resources.

![ob-operator architecture](/img/ob-operator-arch.png)

## Resource Manager

Each resource is uniformly scheduled by its corresponding resource manager, and each resource manager implements the ResourceManager interface, which is defined as follows. It includes methods for resource initialization, resource status updates, resource task retrieval, error message output, and resource deletion operations.

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

The ResourceManager follows a typical state machine model to schedule resources, and the general workflow for resource scheduling is as follows:

1. If it is a new resource, initialize its status field `status.status`.
Retrieve the corresponding `task flow` based on the resource status.
2. If the retrieved task flow is not empty, store it in the `status.operationContext` field of the resource and periodically poll the task status at shorter intervals.
  * If there are pending tasks, submit them to the task manager and set the tasks to `Pending` status while polling their status.
    * If the tasks are successful, proceed to the next task or set the resource to the next state.
    * If the tasks fail, choose to retry or set the resource to an error state.
  * If the retrieved task flow is empty, it indicates that the current resource is running normally without any changes. In this case, re-enqueue the resource with a longer interval.
3. Process and respond to deletion signals for the resource.
4. Update the resource status (including fields like `status.status` and `status.operationContext`).
5. Return the reconciliation result to the ControllerManager, mainly including the re-enqueue interval or error information.

## Task flow and global task manager

Kubernetes internally uses a control loop and message queue to collect and distribute events. Events are dispatched by the Kubernetes control plane to various Controller Managers, which then distribute them to the respective controllers for reconciliation. Each time a controller receives an event, it triggers the reconciliation process. To avoid potential race conditions, the number of worker threads in a controller is typically set to `1`. This means that only one reconciliation task can be started after the previous one is completed. If a reconciliation task takes too long to complete, it may block the reconciliation of other events for the same type of resource. Therefore, the Operator pattern or resource scheduling mode in Kubernetes is generally not suitable for long-running tasks.

To address this issue, ob-operator adopts task flow mechanism and a global task manager to handle long-running tasks. A task flow consists of a list of tasks, the index of the currently executing task, and task status information. The global task manager includes two map structures:

* Workset Map: TaskID -> chan Result, which represents the collection of tasks that are currently executing or have completed but the results have not been read.
* Result Cache Map: TaskID -> Result, which stores the results of completed tasks (success or failure).

The relationship among the control loop, resource manager, and task manager is depicted in the following figure.

![The relationship among the control loop, resource manager, and task manager](/img/ob-operator-task-manager-arch.png)