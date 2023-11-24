# ob-operator 部署

[English version](../en_US/deploy.md) is available.

本文介绍 ob-operator 的部署方式。

## 1. 部署依赖
ob-operator 依赖 [cert-manager](https://cert-manager.io/docs/), cert-manager 的安装可以参考对应的[安装文档](https://cert-manager.io/docs/installation/)

## 2.1 使用 Helm 部署
ob-operator 支持通过 Helm 进行部署，在使用 Helm 命令部署 ob-operator 之前，需要先安装 [Helm](https://github.com/helm/helm)。Helm 安装完成后，可通过如下命令直接部署 ob-operator。

```shell
helm repo add ob-operator https://oceanbase.github.io/ob-operator/
helm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace --version=2.1.0
```

参数说明：

* namespace：命名空间，可自定义，一般建议使用 oceanbase-system。

* version：ob-operator 版本号，建议使用最新的版本。

## 2.2 使用配置文件部署

* Stable
```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/2.1.0_release/deploy/operator.yaml
```
* Development
```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/master/deploy/operator.yaml
```
一般建议使用 Stable 版本的配置文件，如果您想使用开发中的版本，可以选择使用 Development 的配置文件。


## 3. 查看部署结果

部署成功之后可以查看 CRD 的定义。

```shell
kubectl get crds
```

得到如下输出表示部署成功。

```shell
obparameters.oceanbase.oceanbase.com             2023-11-12T08:06:58Z
observers.oceanbase.oceanbase.com                2023-11-12T08:06:58Z
obtenantbackups.oceanbase.oceanbase.com          2023-11-12T08:06:58Z
obtenantrestores.oceanbase.oceanbase.com         2023-11-12T08:06:58Z
obzones.oceanbase.oceanbase.com                  2023-11-12T08:06:58Z
obtenants.oceanbase.oceanbase.com                2023-11-12T08:06:58Z
obtenantoperations.oceanbase.oceanbase.com       2023-11-12T08:06:58Z
obclusters.oceanbase.oceanbase.com               2023-11-12T08:06:58Z
obtenantbackuppolicies.oceanbase.oceanbase.com   2023-11-12T08:06:58Z
```

通过如下命令确认 ob-operator 部署成功。

```shell
kubectl get pods -n oceanbase-system
```

返回结果如下，当看到所有容器都 ready 时并且 status 为 Running， 则表示部署成功。

```shell
NAME                                            READY   STATUS    RESTARTS   AGE
oceanbase-controller-manager-86cfc8f7bf-4hfnj   2/2     Running   0          1m
```
