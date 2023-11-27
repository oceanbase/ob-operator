# ob-operator

ob-operator enables seamless deployment of OceanBase on public cloud or private Kubernetes clusters in the form of containers. 

It provides various functionalities for managing OceanBase clusters, tenants, backup and recovery, and fault recovery. Specifically, ob-operator supports the following features:

- [x] Cluster Management: Bootstrap the cluster, adjust cluster topology, support K8s topology configuration, scale-in/out, cluster upgrade, modify parameters.
- [x] Tenant Management: Create tenants, adjust tenant topology, manage resource units, modify user passwords.
- [x] Backup and Recovery: Periodically backup data to OSS or NFS destinations, restore data from OSS or NFS.
- [x] Physical Standby: Restore standby tenant from backup, create empty standby tenant, activate standby tenant to primary, primary-standby switchover.
- [x] Fault Recovery: Single node fault recovery, cluster-wide fault recovery with IP preservation.

The upcoming features include:

- [ ] Dashboard: A web-based graphical management tool for OceanBase clusters based on ob-operator.
- [ ] Enhanced operational task resources: This includes lightweight tasks focused on cluster and tenant management, among other features.

In the planning phase, there are two additional features:

- [ ] Support for OceanBase Enterprise Edition: The ob-operator is being planned to support the OceanBase Enterprise Edition, which provides additional enterprise-grade features and capabilities.
- [ ] Support for Oracle mode tenants: The ob-operator is also being planned to support Oracle mode tenants, allowing users to run their applications using the Oracle compatibility mode within OceanBase.

## Project Architecture

ob-operator is built on top of kubebuilder and provides control and management of OceanBase clusters and related applications through a unified resource manager interface, a global task manager instance, and a task flow mechanism for handling long-running tasks. The architecture diagram is approximately as follows: 

![ob-operator Architecture](./docs/img/ob-operator-arch.png)

For more detailed information about the architecture, please refer to the [Architecture Document](./docs/en_US/arch.md).

## Requirements

ob-operator is built using the [kubebuilder](https://book.kubebuilder.io/introduction) project, so the development and runtime environment are similar to it.

* To build ob-operator: Go version 1.20 or higher is required.
* To run ob-operator: Kubernetes cluster and kubectl version 1.11.3 or higher are required.
* If using Docker as the container runtime for the cluster, Docker version 17.03 or higher is required.

## Supported OceanBase Versions

ob-operator supports OceanBase v4.x versions. The validated versions include 4.1.x and 4.2.x. It will continue to support new versions of the OceanBase community edition.

OceanBase v3.x versions are currently not supported by ob-operator.

## Quick Start

This section provides a step-by-step guide on how to quickly use ob-operator for managing OceanBase, using examples of deploying ob-operator and deploying an OceanBase cluster.

### Prerequisites

Before getting started, please ensure you have a functional Kubernetes cluster with at least 2 CPU cores, 10GB of memory, and 100GB of storage space. 

ob-operator relies on [cert-manager](https://cert-manager.io/docs/) for certificate management. For instructions on installing cert-manager, please refer to the corresponding [installation](https://cert-manager.io/docs/installation/) documentation. 

OceanBase cluster storage in this example relies on [local-path-provisioner](https://github.com/rancher/local-path-provisioner), which should be installed beforehand.

### Deploy ob-operator

You can deploy ob-operator in a Kubernetes cluster by executing the following command:

* Stable
```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/2.1.0_release/deploy/operator.yaml
```

* Helm chart
```shell
helm repo add ob-operator https://oceanbase.github.io/ob-operator/
helm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace --version=2.1.0
```

* Development
```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/master/deploy/operator.yaml
```

You can verify the successful deployment of ob-operator by executing the following command:

```shell
kubectl get pod -n oceanbase-system

# desired output 
NAME                                            READY   STATUS    RESTARTS   AGE
oceanbase-controller-manager-86cfc8f7bf-4hfnj   2/2     Running   0          1m
```

### Deploy OceanBase Cluster

Before creating an OceanBase cluster, you need to create several secrets to store specific users' passwords for OceanBase.

```shell
kubectl create secret generic root-password --from-literal=password='root_password'
kubectl create secret generic proxyro-password --from-literal=password='proxyro_password'
kubectl create secret generic monitor-password --from-literal=password='monitor_password'
kubectl create secret generic operator-password --from-literal=password='operator_password'
```

You can deploy OceanBase in a Kubernetes cluster by executing the following command:

```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/2.1.0_release/example/quickstart/obcluster.yaml
```

It generally takes around 2 minutes to bootstrap a cluster. Execute the following command to check the status of the cluster. Once the cluster status changes to "running," it indicates that the cluster has been successfully created and bootstrapped:

```shell
kubectl get obclusters.oceanbase.oceanbase.com test

# desired output 
NAME   STATUS    AGE
test   running   6m2s
```

### Connecting to the OceanBase Cluster

Use the following command to find the POD IP of the observer. The naming convention for PODs is {cluster_name}-{cluster_id}-{zone}-uuid:

```shell
kubectl get pods  -o wide
```

To connect, use the following command:

```shell
mysql -h{POD_IP} -P2881 -uroot -proot_password oceanbase -A -c
```

## Documents

Please refer to the [ob-operator documentation](docs/en_US/intro.md) for more information.

## Getting Help

If you encounter any issues while using ob-operator, please feel free to seek help through the following channels:

- [GitHub Issue](https://github.com/oceanbase/ob-operator/issues)
- [Official Website](https://open.oceanbase.com/)

## License

ob-operator is licensed under the [MulanPSL - 2.0](http://license.coscl.org.cn/MulanPSL2) License.
You are free to copy and use the source code. When you modify or distribute the source code, please comply with the MulanPSL - 2.0 Agreement.
