# What is ob-operator
The ob-operator is a Kubernetes operator that simplifies the deployment and management of OceanBase cluster and related resources on Kubernetes.

# Quick Start

## Requirement

### cert-manager

In order to run ob-operator properly, [cert-manager](https://cert-manager.io/docs) needs to be deployed as its dependency. For more details about how to install it, please refer to the [installation](https://cert-manager.io/docs/installation/) document.

Some developers may have trouble accessing images in `quay.io/jetstack` registry. We put an mirrored cert-manager manifest at `deploy/cert-manager.yaml` in which images registry is replaced with `docker.io/oceanbase`, so cert-manager images get more accessible. 

```shell
# Deploy cert-manager with our mirrored images
kubectl apply -f deploy/cert-manager.yaml
```

The mirrored manifests of cert-manager are of version `12.0.4`. If there is any warning thrown out when applying `deploy/cert-manager.yaml`, please apply another compatible version.

### local-path-provisioner

Cluster in the quick start uses [local-path-provisioner](https://github.com/rancher/local-path-provisioner) as PVC provisioner. If you keep the storage class `local-path`, this component is required.

```shell
# Apply local-path-provisioner to k8s cluster
kubectl apply -f https://raw.githubusercontent.com/rancher/local-path-provisioner/v0.0.25/deploy/local-path-storage.yaml
```

## Deploy ob-operator

### Using helm
[Helm](https://github.com/helm/helm) is a package management tool for Kubernetes, please refer to the helm documentation to install the helm client.

```shell
helm repo add ob-operator https://oceanbase.github.io/ob-operator/
helm install ob-operator ob-operator/ob-operator --namespace=oceanbase-system --create-namespace --version=2.1.0
```

### Using configuration file

The configuration files are located under deploy directory, using the following commands to deploy ob-operator.

```shell
# Deploy ob-operator
kubectl apply -f deploy/operator.yaml
```

## Deploy OceanBase cluster

### Customize configuration file

`obcluster.yaml` in `example/obcluster` defines an OceanBase cluster, including deployment topology, resources, storages etc. You can configure your own OceanBase cluster based on this file.

### Deploy OceanBase cluster

Create namespace if needed, namespace should match the one in configuration file `obcluster.yaml`.

```shell
kubectl create namespace oceanbase
```

Create secret for users, secret name must be the same as these configured in `obcluster.yaml` under `spec.userSecrets`.

```shell
# create secret to hold password for user root
kubectl create secret -n oceanbase generic root-password --from-literal=password='******'

# create secret to hold password for user proxyro, proxyro is a readonly user for obproxy to query meta info
kubectl create secret -n oceanbase generic proxyro-password --from-literal=password='******'

# create secret to hold password for user monitor, monitor is a readonly user for obagent to query metric data
kubectl create secret -n oceanbase generic monitor-password --from-literal=password='******'

# create secret to hold password for user operator, operator is the admin user for obproxy to maintain obcluster
kubectl create secret -n oceanbase generic operator-password --from-literal=password='******'
```

Using the following command to deploy OceanBase cluster.

```shell
kubectl apply -f example/obcluster/obcluster.yaml
```

It may take a while to complete the whole process to deploy OceanBase cluster, you can use the following command to check whether it's finished. It may cost 2~3 minutes to bootstrap the cluster in usual.

```shell
# use kubectl get
kubectl get obclusters test -n oceanbase -o yaml
# or use kubectl wait
kubectl wait -n oceanbase obclusters test --for=jsonpath='{.status.status}'=running --timeout=10m
```
Wait until the status of OBCluster resource turns into `running`.

### Connect to OceanBase Cluster
After successfully deploying OceanBase cluster, you can connect to OceanBase cluster via PodIP of any observer pod.

```shell
# connect the root user of sys tenant
mysql -h{POD_IP} -P2881 -uroot -p${ROOT_PWD} oceanbase -A -c
```

# Contributing
Contributions are warmly welcomed and greatly appreciated. Here are a few ways you can contribute:
- Raise us an [Issue](https://github.com/oceanbase/ob-operator/issues).
- Create a [Pull Request](https://github.com/oceanbase/ob-operator/pulls).

# License
Ob-operator is licensed under the [MulanPSL - 2.0](http://license.coscl.org.cn/MulanPSL2) license. You can copy and use the source code freely. When you modify or distribute the source code, please follow the MulanPSL - 2.0 license.
