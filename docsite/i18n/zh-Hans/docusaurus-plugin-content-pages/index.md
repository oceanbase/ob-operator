---
title: 项目介绍
---


# ob-operator

`ob-operator` 是满足 Kubernetes Operator 扩展范式的自动化工具，可以极大简化在 Kubernetes 上部署和管理 OceanBase 集群及相关资源的过程。

## 快速上手

这部分以一个简单示例说明如何使用 ob-operator 快速部署 OceanBase 集群。

### 前提条件

开始之前请准备一套可用的 Kubernetes 集群，并且至少可以分配 2C, 10G 内存以及 100G 存储空间。

ob-operator 依赖 [cert-manager](https://cert-manager.io/docs/), cert-manager 的安装可以参考对应的[安装文档](https://cert-manager.io/docs/installation/)，如果您无法访问官方制品托管在 `quay.io` 镜像站的镜像，可通过下面的指令安装我们转托在 `docker.io` 中的制品：

```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/stable/deploy/cert-manager.yaml
```

本例子中的 OceanBase 集群存储依赖 [local-path-provisioner](https://github.com/rancher/local-path-provisioner) 提供, 需要提前进行安装并确保其存储目的地有足够大的磁盘空间。如果您计划在生产环境部署，推荐使用其他的存储解决方案。我们在[存储兼容性](#存储兼容性)一节提供了我们测试过的存储兼容性结果。

### 部署 ob-operator

#### 使用 YAML 配置文件·

通过以下命令即可在 K8s 集群中部署 ob-operator：

- 稳定版本

```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/stable/deploy/operator.yaml
```

- 开发版本

```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/master/deploy/operator.yaml
```

#### 使用 Helm Chart

Helm Chart 将 ob-operator 部署的命名空间进行了参数化，可在安装 ob-operator 之前指定命名空间。

```shell
helm repo add ob-operator https://oceanbase.github.io/ob-operator/
helm repo update
helm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace
```

#### 使用 terraform

部署所需要的文件放在项目的 `deploy/terraform` 目录

1. 生成配置变量:
   在开始部署前，需要通过以下命令来生成 `terraform.tfvars` 文件，用来记录当前 Kubernetes 集群的一些配置。

```shell
cd deploy/terraform
./generate_k8s_cluster_tfvars.sh
```

2. 初始化 Terraform:
   此步骤用来保证 terraform 获取到必要的 plugin 和模块来管理配置的资源，使用如下命令来进行初始化。

```
terraform init
```

3. 应用配置:
   执行以下命令开始部署 ob-operator。

```
terraform apply
```

#### 验证部署结果

安装完成之后，可以使用以下命令验证 ob-operator 是否部署成功：

```shell
kubectl get pod -n oceanbase-system

# 预期的输出
NAME                                            READY   STATUS    RESTARTS   AGE
oceanbase-controller-manager-86cfc8f7bf-4hfnj   2/2     Running   0          1m
```

### 部署 OceanBase 集群

创建 OceanBase 集群之前，需要先创建好若干 secret 来存储 OceanBase 中的特定用户的密码：

```shell
kubectl create secret generic root-password --from-literal=password='root_password'
```

通过以下命令即可在 K8s 集群中部署 OceanBase：

```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/stable/example/quickstart/obcluster.yaml
```

一般初始化集群需要 2 分钟左右的时间，执行以下命令，查询集群状态，当集群状态变成 running 之后表示集群创建和初始化成功：

```shell
kubectl get obclusters.oceanbase.oceanbase.com test

# desired output
NAME   STATUS    AGE
test   running   6m2s
```

### 连接集群

通过以下命令查找 observer 的 POD IP，POD 名的规则是 `${cluster_name}-${cluster_id}-${zone}-uuid`：

```shell
kubectl get pods -o wide
```

通过以下命令连接：

```shell
mysql -h{POD_IP} -P2881 -uroot -proot_password oceanbase -A -c
```

### OceanBase Dashboard

我们很高兴向用户推出创新的 OceanBase Kubernetes Dashboard，这是一款旨在改善用户在 Kubernetes 上管理和监控 OceanBase 集群体验的先进工具。欢迎各位用户使用和反馈，同时我们也在积极开发新功能以增强未来的更新。[快速上手](/docs/dashboard/quickstart)文档能帮助您快速了解 OceanBase 的功能和使用方法。

安装 OceanBase Dashboard 非常简单, 只需要执行如下命令。

```
helm repo add ob-operator https://oceanbase.github.io/ob-operator/
helm repo update ob-operator
helm install oceanbase-dashboard ob-operator/oceanbase-dashboard
```

![oceanbase-dashboard-install](/img/oceanbase-dashboard-install.jpg)

OceanBase Dashboard 成功安装之后, 会自动创建一个 admin 用户和随机密码，可以通过如下命令查看密码。

```
echo $(kubectl get -n default secret oceanbase-dashboard-user-credentials -o jsonpath='{.data.admin}' | base64 -d)
```

一个 NodePort 类型的 service 会默认创建，可以通过如下命令查看 service 的地址，然后在浏览器中打开。

```
kubectl get svc oceanbase-dashboard-oceanbase-dashboard
```

![oceanbase-dashboard-service](/img/oceanbase-dashboard-service.jpg)

使用 admin 账号和查看到的密码登录。

![oceanbase-dashboard-overview](/img/oceanbase-dashboard-overview.jpg)
![oceanbase-dashboard-topology](/img/oceanbase-dashboard-topology.jpg)

## 项目架构

ob-operator 以 kubebuilder 为基础，通过统一的资源管理器接口、全局的任务管理器实例以及解决长调度的任务流机制完成对 OceanBase 集群及相关应用的控制和管理。ob-operator 的架构大致如下图所示：

![ob-operator 架构设计](/img/ob-operator-arch.png)

![ob-operator 任务管理器](/img/ob-operator-task-manager-arch.png)

有关架构细节可参见[架构设计文档](/docs/developer/arch)。

## 特性

ob-operator 支持 OceanBase 集群的管理、租户管理、备份恢复、故障恢复等功能，具体而言支持了以下功能：

- [x] 集群管理：集群自举、调整集群拓扑、支持 K8s 拓扑配置、扩缩容、集群升级、修改参数
- [x] 租户管理：创建租户、调整租户拓扑、管理资源单元、修改用户密码
- [x] 备份恢复：向 OSS 或 NFS 目的地周期性备份数据、从 OSS 或 NFS 中恢复数据
- [x] 物理备库：从备份中恢复出备租户、创建空备租户、备租户升主、主备切换
- [x] 故障恢复：单节点故障恢复，IP 保持情况下的集群故障恢复
- [x] Dashboard(GUI)：基于 ob-operator 的图形化 OceanBase 集群管理工具

## 存储兼容性

我们测试了如下的存储方案，兼容性结果如表格所示：

| 存储方案               | 测试版本 | 是否兼容 | 说明                               |
| ---------------------- | -------- | -------- | ---------------------------------- |
| local-path-provisioner | 0.0.23   | ✅       | 建议开发和测试环境使用             |
| Rook CephFS            | v1.6.7   | ❌       | CephFS 不支持 `fallocate` 系统调用 |
| Rook RBD (Block)       | v1.6.7   | ✅       |                                    |
| OpenEBS (cStor)        | v3.6.0   | ✅       |                                    |
| GlusterFS              | v1.2.0   | ❓       | 要求机器内核版本不低于 5.14        |
| Longhorn               | v1.6.0   | ✅       |                                    |
| JuiceFS                | v1.1.2   | ✅       |                                    |
| NFS                    | v5.5.0   | ❌       | NFS 协议 >= 4.2 时能启动集群，但无法回收租户资源            |

## 支持的 OceanBase 版本

ob-operator 支持 OceanBase v4.x 版本。某些特性需要特定的 OceanBase 版本，可参考用户手册获取详细信息。

暂不支持 OceanBase v3.x 版本。

## 环境依赖

ob-operator 使用 [kubebuilder](https://book.kubebuilder.io/introduction) 项目进行构建，所以开发和运行环境与其相近。

- 构建 ob-operator 需要 Go 1.22 版本及以上；
- 运行 ob-operator 需要 Kubernetes 集群和 kubectl 的版本在 1.18 及以上。我们在 1.23 ~ 1.28 版本的 K8s 集群上检验过 ob-operator 的运行是符合预期的。
- 如果使用 Docker 作为集群的容器运行时，需要 Docker 17.03 及以上版本；我们的构建和运行环境使用的 Docker 版本为 18。

## 文档

- [ob-operator 架构设计](/docs/developer/arch)
- [部署 ob-operator](/docs/developer/deploy)
- [开发手册](/docs/developer/contributor-guidance)
- [用户手册](/docs/manual/what-is-ob-operator)

## 获取帮助

如果您在使用 ob-operator 时遇到任何问题，欢迎通过以下方式寻求帮助：

- [GitHub Issue](https://github.com/oceanbase/ob-operator/issues)
- [官方论坛](https://ask.oceanbase.com/)
- [Slack](https://oceanbase.slack.com/archives/C053PT371S7)
- 微信群（请添加小助手微信，微信号: OBCE666）
- 钉钉群（[二维码](/img/dingtalk-operator-users.png)）

<div style={{ width: '200px' }}>
  ![钉钉群二维码](/img/dingtalk-operator-users.png)
</div>

## 参与开发

- [提出 Issue](https://github.com/oceanbase/ob-operator/issues)
- [发起 Discussion](https://github.com/oceanbase/ob-operator/discussions)
- [发起 Pull request](https://github.com/oceanbase/ob-operator/pulls)

## 许可证

ob-operator 使用 [MulanPSL - 2.0](http://license.coscl.org.cn/MulanPSL2) 许可证。
您可以免费复制及使用源代码。当您修改或分发源代码时，请遵守木兰协议。
