# 自定义配置

## 部署 CRD

```yaml
kubectl apply -f ./deploy/crd.yaml
```

## 部署 ob-operator

您需要根据您的配置修改 `operator.yaml` 。

您需要添加启动参数 `--cluster-name`，该参数需要与 obcluster 中的 `cluster` 配置一致。
该配置的含义：ob-operator 只会处理 `cluster` 的值与自身启动参数 `--cluster-name` 的值相同的 CRD。

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  labels:
    control-plane: controller-manager
  name: ob-operator-controller-manager
  namespace: oceanbase-system
spec:
  replicas: 1
  selector:
    matchLabels:
      control-plane: controller-manager
  template:
    metadata:
      labels:
        control-plane: controller-manager
    spec:
      containers:
      - args:
        - --health-probe-bind-address=:8081
        - --metrics-bind-address=127.0.0.1:8080
        - --leader-elect
        - --cluster-name=cn
        command:
        - /manager
        image: ob-operator:latest
        imagePullPolicy: Always
        name: manager
```

## 配置节点 label

需要将 Kubernetes 节点打 label，label 需要与 obcluster.yaml 中 `nodeSelector` 配置相匹配。ob-operator 会将 Pod 调度到具有相应 label 的节点上。

推荐配置 label 的 key 为 `topology.kubernetes.io/zone`。不同 Zone 推荐配置不同的 label 以做容灾。

```yaml
kubectl label node nodename topology.kubernetes.io/zone=zonename
```

## 部署 OceanBase 集群

`obcluster.yaml` 中需要用户根据自己的配置做一些修改。

```yaml
apiVersion: cloud.oceanbase.com/v1
kind: OBCluster
metadata:
  namespace: oceanbase
  name: ob-test
spec:
  version: 3.1.1-4
  clusterID: 1
  topology:
    - cluster: cn
      zone:
      - name: zone1
        region: regio1
        nodeSelector:
          topology.kubernetes.io/zone: zone1
        replicas: 1
      - name: zone2
        region: regio1
        nodeSelector:
          topology.kubernetes.io/zone: zone2
        replicas: 1
      - name: zone3
        region: regio1
        nodeSelector:
          topology.kubernetes.io/zone: zone3
        replicas: 1
  resources:
    cpu: 2
    memory: 10Gi
    storage:
      - name: data-file
        storageClassName: "local-path"
        size: 50Gi
      - name: data-log
        storageClassName: "local-path"
        size: 50Gi
      - name: log
        storageClassName: "local-path"
        size: 30Gi
```

- `version` OceanBase 集群的版本。
- `cluster` 需要按需配置，如果需要在该 Kubernetes 集群中部署 OceanBase 集群，请将 `cluster` 配置为与 ob-operator 启动参数 `--cluster-name` 相同的配置。
- `cpu` 配置建议为大于 2 的整数，小于 2 会引发系统异常。
- `memory` 配置建议为大于 10Gi 的整数，小于 10Gi 会引发系统异常。
- `storage` 的 `data-file` 部分为 OBServer 系统配置项 `datafile_size` 的大小，建议为 `memory` 的 3 倍以上。`storageClassName` 用户可以自行按需配置。
- `storage` 的 `data-log` 部分为 OBServer 系统配置项 `data_dir` 的大小，建议为 `memory` 的 5 倍以上。`storageClassName` 用户可以自行按需配置。
- `storage` 的 `log` 部分为 OBServer 系统日志的大小，建议为 30Gi 以上。`storageClassName` 用户可以自行按需配置。

`nodeSelector` 的数据结构：

```go
// NodeSelector is a selector which must be true for the pod to fit on a node.
// Selector which must match a node's labels for the pod to be scheduled on that node.
// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
// +optional
// +mapType=atomic
NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`
```
