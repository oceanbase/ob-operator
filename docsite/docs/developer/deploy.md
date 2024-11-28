# Deploy ob-operator 

This article introduces the deployment methods for ob-operator.

## 1. Deployment Dependencies

ob-operator relies on [cert-manager](https://cert-manager.io/docs/). You can refer to the corresponding installation documentation for the [installation of cert-manager](https://cert-manager.io/docs/installation/).

## 2.1 Deploying with Helm

ob-operator supports deployment using Helm. Before deploying ob-operator with the Helm command, you need to install [Helm](https://github.com/helm/helm). After Helm is installed, you can deploy ob-operator directly using the following command.

```shell
helm repo add ob-operator https://oceanbase.github.io/ob-operator/
helm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace
```

Parameters:

* namespace: Namespace, can be customized. It is recommended to use "oceanbase-system" as the namespace.

* version: ob-operator version number. It is recommended to use the latest version `2.3.1`.

## 2.2 Deploying with Configuration Files

* Stable
```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/2.3.1_release/deploy/operator.yaml
```
* Development
```shell
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/master/deploy/operator.yaml
```

It is generally recommended to use the configuration files for the stable version. However, if you want to use a development version, you can choose to use the configuration files for the development version.

## 3. Check the deployment results

After a successful deployment, you can view the definition of Custom Resource Definitions (CRDs) by executing the following command:

```shell
kubectl get crds
```

If you get the following output, it indicates a successful deployment:

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

To confirm whether ob-operator has been successfully deployed, you can use the following command:

```shell
kubectl get pods -n oceanbase-system
```

The result will look like the following example. If you see that all containers are ready and the status is "Running", it indicates a successful deployment.

```shell
NAME                                            READY   STATUS    RESTARTS   AGE
oceanbase-controller-manager-86cfc8f7bf-4hfnj   2/2     Running   0          1m
```
