# What is ob-operator
The ob-operator is a Kubernetes operator that simplifies the deployment and management of OceanBase cluster and related resources on Kubernetes.

# Quick Start
## Requirement
In order to run ob-operator properly, [cert-manager](https://cert-manager.io/docs) needs to be deployed as its dependency, for more details about how to install it, please refer to the [installation](https://cert-manager.io/docs/installation/) document.

## Deploy ob-operator
### Using helm
[Helm](https://github.com/helm/helm) is a package management tool for Kubernetes, please refer to the helm documentation to install the helm client.

```
helm repo add ob-operator https://oceanbase.github.io/ob-operator/
helm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace  --version=2.0.0
```

### Using configuration file
The configuration files are located under deploy directory, using the following commands to deploy ob-operator.
```
# Deploy ob-operator
kubectl apply -f deploy/operator.yaml
```

## Deploy OceanBase cluster
### Customize configuration file
`deploy/obcluster.yaml` defines an OceanBase cluster, including deployment topology, resources, storages etc. You can configure your own OceanBase based on this file.

### Deploy OceanBase
Create namespace if needed, namespace should match the one in configuration file `deploy/obcluster.yaml`
```
kubectl create namespace ${namespace_name}
```
Using the following command to deploy OceanBase Cluster
```
kubectl apply -f deploy/obcluster.yaml
```
It may take a while to complete the whole process to deploy OceanBase cluster, you can use the following command to check whether it's finished
```
kubectl get obclusters ${name} -n ${namespace} -o yaml
```
wait until the status of obclster resource turns into running.


### Connect to OceanBase Cluster
After successfully deployed OceanBase cluster, you can connect to OceanBase cluster via any observer pod's ip.

# Contributing
Contributions are warmly welcomed and greatly appreciated. Here are a few ways you can contribute:
- Raise us an [Issue](https://github.com/oceanbase/ob-operator/issues).
- Create a [Pull Request](https://github.com/oceanbase/ob-operator/pulls).

# License
Ob-operator is licensed under the [MulanPSL - 2.0](http://license.coscl.org.cn/MulanPSL2) license. You can copy and use the source code freely. When you modify or distribute the source code, please follow the MulanPSL - 2.0 license.
