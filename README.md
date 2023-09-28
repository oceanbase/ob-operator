# New Release
We have the latest 2.0.0 version released, please refer to branch [2.0.0_release](https://github.com/oceanbase/ob-operator/tree/2.0.0_release)

# What is ob-operator
The ob-operator is a Kubernetes operator that simplifies the deployment and management of OceanBase clusters on Kubernetes.

# Quick Start
## Deploy ob-operator
### Using Helm
[Helm](https://github.com/helm/helm) is a package management tool for Kubernetes, please refer to the helm documentation to install the helm client.

```
helm repo add ob-operator https://oceanbase.github.io/ob-operator/
helm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace  --version=1.1.0
```

### Using configuration file
The configuration files are located under deploy directory, using the following commands to deploy ob-operator.
```
# Deploy CRD
kubectl apply -f deploy/crd.yaml
# Deploy ob-operator
kubectl apply -f deploy/operator.yaml
```

## Deploy OceanBase cluster
### Customize configuration file
`deploy/obcluster.yaml` defines an OceanBase cluster, including deployment topology, resources, storages etc. You can configure your own OceanBase based on this file.


### Label node
Ob-operator distributes observer pods on selected nodes to achive high availablity, by matching specific labels, ob-operator knows where to locate an observer pod, the node labels must includes the labels configured under nodeSelector in `deploy/obcluster.yaml`.
You can use the following command to label a node.
```
kubectl label node ${node_name} ${label_name}=${label_value}
```

### Deploy OceanBase
Create namespace if needed, namespace should match the one in configuration file `deploy/obcluster.yaml`
```
kubectl create namespace ${namespace_name}
```
Using the following command to deploy OceanBase Cluster
```
kubectl apply -f deploy/obcluster.yaml
```

### Connect to OceanBase Cluster
After successfully deployed OceanBase cluster, you can connect to OceanBase cluster via any observer pod's ip or the service address created by ob-operator named `svc-${cluster_name}`, `${cluster_name}` is the name which you configured in `deploy/obcluster.yaml`

# Contributing
Contributions are warmly welcomed and greatly appreciated. Here are a few ways you can contribute:
- Raise us an [Issue](https://github.com/oceanbase/ob-operator/issues).
- Create a [Pull Request](https://github.com/oceanbase/ob-operator/pulls).

# License
Ob-operator is licensed under the [MulanPSL - 2.0](http://license.coscl.org.cn/MulanPSL2) license. You can copy and use the source code freely. When you modify or distribute the source code, please follow the MulanPSL - 2.0 license.
