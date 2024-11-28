#!/bin/bash

set -xe

CLUSTER_NAME=openstack
TENANT_NAME=openstack
OCEANBASE_NAMESPACE=openstack
STORAGE_CLASS=local-path
OCEANBASE_CLUSTER_IMAGE=oceanbase/oceanbase-cloud-native:4.2.5.0-100000052024102022
OBPROXY_IMAGE=oceanbase/obproxy-ce:4.3.2.0-26
ROOT_PASSWORD=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 16 | base64 -w 0)
PROXYRO_PASSWORD=$(head /dev/urandom | tr -dc A-Za-z0-9 | head -c 16 | base64 -w 0)

# check and install cert-manager
if kubectl get crds -o name | grep 'certificates.cert-manager.io'
then
    echo "cert-manager is already installed"
else
    echo "cert-manager is not installed, install it now"
    kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/refs/heads/master/deploy/cert-manager.yaml
fi

# install ob-operator, always apply the newest version
echo "install or update ob-operator"
kubectl apply -f https://raw.githubusercontent.com/oceanbase/ob-operator/refs/heads/2.3.1_release/deploy/operator.yaml
kubectl wait --for=condition=Ready pod -l "control-plane=controller-manager" -n oceanbase-system --timeout=300s

tee /tmp/namespace.yaml <<EOF
apiVersion: v1
kind: Namespace
metadata:
  name: ${OCEANBASE_NAMESPACE}
EOF

kubectl apply -f /tmp/namespace.yaml

tee /tmp/secret.yaml <<EOF
apiVersion: v1
kind: Secret
metadata:
  name: root-password
  namespace: ${OCEANBASE_NAMESPACE}
type: Opaque
data:
  password: ${ROOT_PASSWORD}
---
apiVersion: v1
kind: Secret
metadata:
  name: proxyro-password
  namespace: ${OCEANBASE_NAMESPACE}
type: Opaque
data:
  password: ${PROXYRO_PASSWORD}
EOF

kubectl apply -f /tmp/secret.yaml

tee /tmp/obcluster.yaml <<EOF
apiVersion: oceanbase.oceanbase.com/v1alpha1
kind: OBCluster
metadata:
  name: ${CLUSTER_NAME}
  namespace: ${OCEANBASE_NAMESPACE}
  annotations:
    "oceanbase.oceanbase.com/mode": "service"
spec:
  clusterName: ${CLUSTER_NAME}
  clusterId: 1
  serviceAccount: "default"
  userSecrets:
    root: root-password
    proxyro: proxyro-password
  topology:
    - zone: zone1
      replica: 1
    - zone: zone2
      replica: 1
    - zone: zone3
      replica: 1
  observer:
    image: ${OCEANBASE_CLUSTER_IMAGE}
    resource:
      cpu: 2
      memory: 10Gi
    storage:
      dataStorage:
        storageClass: ${STORAGE_CLASS}
        size: 50Gi
      redoLogStorage:
        storageClass: ${STORAGE_CLASS}
        size: 50Gi
      logStorage:
        storageClass: ${STORAGE_CLASS}
        size: 20Gi
  parameters:
    - name: system_memory
      value: 2G
    - name: __min_full_resource_pool_memory
      value: "2147483648"
EOF

kubectl apply -f /tmp/obcluster.yaml
kubectl wait --for=jsonpath='{.status.status}'=running obcluster/${CLUSTER_NAME} -n ${OCEANBASE_NAMESPACE}  --timeout=900s

tee /tmp/obtenant.yaml <<EOF
apiVersion: oceanbase.oceanbase.com/v1alpha1
kind: OBTenant
metadata:
  name: ${TENANT_NAME}
  namespace: ${OCEANBASE_NAMESPACE}
spec:
  obcluster: ${CLUSTER_NAME}
  tenantName: ${TENANT_NAME}
  variables:
    - name: autocommit
      value: "0"
    - name: version
      value: "8.0.30"
  unitNum: 1
  charset: utf8mb4
  connectWhiteList: '%'
  forceDelete: true
  credentials:
    root: root-password
  pools:
    - zone: zone1
      type:
        name: Full
        replica: 1
        isActive: true
      resource:
        maxCPU: 2
        memorySize: 4Gi
    - zone: zone2
      type:
        name: Full
        replica: 1
        isActive: true
      resource:
        maxCPU: 2
        memorySize: 4Gi
    - zone: zone3
      type:
        name: Full
        replica: 1
        isActive: true
      priority: 3
      resource:
        maxCPU: 2
        memorySize: 4Gi
EOF

kubectl apply -f /tmp/obtenant.yaml
kubectl wait --for=jsonpath='{.status.status}'=running obtenant/${TENANT_NAME} -n ${OCEANBASE_NAMESPACE}  --timeout=300s

RS_LIST=$(kubectl get observers -l ref-obcluster=${CLUSTER_NAME} -n ${OCEANBASE_NAMESPACE} -o jsonpath='{range .items[*]}{.status.serviceIp}{":2881;"}' | sed 's/;:2881;$//g')
echo $RS_LIST

tee /tmp/obproxy.yaml <<EOF
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm-obproxy-${CLUSTER_NAME}
  namespace: ${OCEANBASE_NAMESPACE}
data:
  ODP_MYSQL_VERSION: 8.0.30
  ODP_PROXY_TENANT_NAME: ${TENANT_NAME}
---
apiVersion: v1
kind: Service
metadata:
  name: svc-obproxy-${CLUSTER_NAME}
  namespace: ${OCEANBASE_NAMESPACE}
spec:
  ports:
  - name: sql
    port: 2883
    protocol: TCP
    targetPort: 2883
  - name: prometheus
    port: 2884
    protocol: TCP
    targetPort: 2884
  selector:
    app: app-obproxy-${CLUSTER_NAME}
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: obproxy-${CLUSTER_NAME}
  namespace: ${OCEANBASE_NAMESPACE}
spec:
  replicas: 2
  selector:
    matchLabels:
      app: app-obproxy-${CLUSTER_NAME}
  template:
    metadata:
      labels:
        app: app-obproxy-${CLUSTER_NAME}
    spec:
      containers:
      - env:
        - name: APP_NAME
          value: obproxy-${CLUSTER_NAME}
        - name: OB_CLUSTER
          value: ${CLUSTER_NAME}
        - name: RS_LIST
          value: ${RS_LIST}
        - name: PROXYRO_PASSWORD
          valueFrom:
            secretKeyRef:
              key: password
              name: proxyro-password
        envFrom:
        - configMapRef:
            name: cm-obproxy-${CLUSTER_NAME}
        image: ${OBPROXY_IMAGE}
        imagePullPolicy: IfNotPresent
        name: obproxy
        ports:
        - containerPort: 2883
          name: sql
          protocol: TCP
        - containerPort: 2884
          name: prometheus
          protocol: TCP
        resources:
          limits:
            cpu: "1"
            memory: 2Gi
          requests:
            cpu: "1"
            memory: 2Gi
EOF

kubectl apply -f /tmp/obproxy.yaml
kubectl wait --for=condition=Ready pod -l app="app-obproxy-${CLUSTER_NAME}" -n ${OCEANBASE_NAMESPACE} --timeout=300s

echo "OceanBase is ready, you may use the following connection"
echo "mysql -hsvc-obproxy-${CLUSTER_NAME}.${OCEANBASE_NAMESPACE}.svc -P2883 -uroot -p$(echo ${ROOT_PASSWORD} | base64 -d)"
