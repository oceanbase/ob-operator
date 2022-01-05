# e2e Test

The test requires a real Kubernetes environment.
The test suite deploys the application in a Kubernetes cluster and tests whether the application meets expectations.

## Deploy ob-operator

The operator startup parameter `--cluster-name` must be the same as that configured in YAML used for the test.

The following uses YAML to deployment operator as an example:

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
        - --cluster-name=test  # the value must be the same as that configured in YAML used for the test, The default value is cn
        command:
        - /manager
        image: ob-operator:latest
        imagePullPolicy: Always
        name: manager
```

The network that requires the test environment can communicate with the internal network of the Kubernetes cluster or run tests within the Kubernetes cluster.
Tests in the observer process access the ClusterIP of the Service created for the observer.

## Test for statefulapp

```
ginkgo --focus=/statefulapp/ --regexScansFilePath=true --slowSpecThreshold=3600 --progress --reportPassed
```

## Test for observer

```
ginkgo --focus=/observer/ --regexScansFilePath=true --slowSpecThreshold=3600 --progress --reportPassed
```
