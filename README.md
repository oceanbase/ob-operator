# What is ob-operator
The ob-operator is a Kubernetes operator that simplifies the deployment and management of OceanBase clusters on Kubernetes.

# Quick Start
## Installation
### Using Helm
Helm is a package management tool for Kubernetes, Refer to Helm official documentation to install the Helm client.
```
helm repo add ob-operator https://oceanbase.github.io/ob-operator/
helm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace  --version=1.1.0
```

### Using configuration file
The configuration files are located at deploy directory, using the following commands to deploy ob-operator.
```
# Deploy CRD
kubectl apply -f deploy/crd.yaml
# Deploy ob-operator
kubectl apply -f deploy/operator.yaml
```

## Deploy OceanBase
### Customize configuration file
`deploy/obcluster.yaml` defines an OceanBase cluster, including deployment topology, resources etc. You can configure your own OceanBase based on this file.


### Label node
Ob-operator distributes observer pods on selected nodes to achive high availablity, by matching specific labels, ob-operator knows where to locate an observer pod, the node labels must includes the labels configured under nodeSelector in `deploy/obcluster.yaml`.
You can use the following command to label a node.
```
kubectl label node ${node_name} ${label_name}=${label_value}
```

### Deploy OceanBase
Using the following command to deploy OceanBase Cluster
```
kubectl apply -f deploy/obcluster.yaml
```

### Connect to OceanBase Cluster
After successfully deployed OceanBase cluster, you can connect to OceanBase cluster via any observer pod's ip or the service address created by ob-operator named svc-${cluster_name}

# Contributing
We highly appreciate any kinds of contribution. For more details, please refer to How to contribute

# License
Ob-operator is licensed under the Mulan Public License, Version 2. See the LICENSE file for more info.





