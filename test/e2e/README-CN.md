# e2e 测试

测试需要真实的 Kubernetes 环境。
测试套件会在 Kubernetes 集群中部署应用，并测试应用是否达到预期。

## 部署 ob-operator

其中，operator 启动参数 `--cluster-name` 需要与测试所使用的 YAML 中配置的保持一致。

以 YAML 方式部署为例：

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
        - --cluster-name=test  # 需要与测试 YAML 中配置的保持一致，默认是 cn
        command:
        - /manager
        image: ob-operator:latest
        imagePullPolicy: Always
        name: manager
```

需要测试环境的网络与 Kubernetes 集群内部网络互通，或者在 Kubernetes 集群内跑测试。
observer 阶段的测试会访问 observer 的 Service ClusterIP。

## 测试 statefulapp

```
ginkgo --focus=/statefulapp/ --regexScansFilePath=true --slowSpecThreshold=3600 --progress --reportPassed
```

## 测试 observer

```
ginkgo --focus=/observer/ --regexScansFilePath=true --slowSpecThreshold=3600 --progress --reportPassed
```
