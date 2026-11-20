# OceanBase Cluster Chart

Install the OceanBase Operator before using this chart. The Lakehouse
(`shared_storage`) mode requires an Operator release whose CRDs include both
`OBCluster.spec.deploymentMode/sharedStorageInfo/logServiceRef` and the
`OBLogServiceCluster` custom resource. The Operator commit `fd0f9362` and later
releases include these APIs.

Before running Helm, verify that `kubectl get nodes` reaches the intended
cluster. On K3s, use the same kubeconfig for kubectl, Helm, and the progress
script. If the default configuration is unavailable, set `KUBECONFIG` to
a [private kubeconfig you can read](https://docs.k3s.io/cluster-access),
such as `$HOME/.kube/k3s.yaml`.

For a new Lakehouse environment using K3s 1.35, run
`stat -fc %T /sys/fs/cgroup` on each node and check for `cgroup2fs`.
Use an upstream DNS resolver reachable from the nodes that resolves the image
registry. Avoid setting the host resolver to the cluster's CoreDNS service IP,
which can create a [DNS forwarding loop](https://coredns.io/plugins/loop/#troubleshooting).

## Traditional mode

By default, this chart deploys a normal `OBCluster` and omits Lakehouse
resources.

```bash
helm install oceanbase . -n oceanbase --create-namespace
```

## Lakehouse mode with generated S3 secrets

Set two different buckets and provide credentials once. The chart creates two
Secrets with the keys required by Operator: `access_id` and `access_key`.
LogService bootstrap needs three distinct zones; the Lakehouse defaults create
one replica in each of `zone1`, `zone2`, and `zone3`. Schedule those replicas on
separate failure domains for production.

For in-cluster MinIO, use its Service IP as the S3 endpoint. The LogService S3
client uses bucket-prefixed DNS names for hostname endpoints, which the normal
Kubernetes Service DNS does not provide. An external S3 endpoint with wildcard
bucket DNS also works. For the examples below, obtain the IP with:

```bash
MINIO_CLUSTER_IP=$(kubectl get service minio -n lakehouse -o jsonpath='{.spec.clusterIP}')
```

```bash
helm install lakehouse . \
  --namespace lakehouse --create-namespace \
  --set lakehouse.enabled=true \
  --set lakehouse.observerImage=<oceanbase-ai-image> \
  --set lakehouse.logService.image=<oblogservice-image> \
  --set "lakehouse.logService.bucketURL=s3://lakehouse-logservice?host=http://${MINIO_CLUSTER_IP}:9000&s3_region=us-east-1" \
  --set "lakehouse.sharedStorage.bucketURL=s3://lakehouse-data?host=http://${MINIO_CLUSTER_IP}:9000&s3_region=us-east-1" \
  --set lakehouse.credentials.accessId=<access-id> \
  --set lakehouse.credentials.accessKey=<access-key>
```

## Lakehouse mode with existing S3 secrets

Create one or two Secrets in the release namespace before installation. Each
referenced Secret must contain non-empty `access_id` and `access_key`; this
chart does not consume the `s3Region` key used by backup-policy Secrets.

```bash
kubectl create secret generic existing-logservice-s3 \
  --from-literal=access_id=<access-id> \
  --from-literal=access_key=<access-key> \
  -n lakehouse

helm install lakehouse . \
  --namespace lakehouse --create-namespace \
  --set lakehouse.enabled=true \
  --set lakehouse.observerImage=<oceanbase-ai-image> \
  --set lakehouse.logService.image=<oblogservice-image> \
  --set "lakehouse.logService.bucketURL=s3://lakehouse-logservice?host=http://${MINIO_CLUSTER_IP}:9000&s3_region=us-east-1" \
  --set lakehouse.logService.existingSecretName=existing-logservice-s3 \
  --set "lakehouse.sharedStorage.bucketURL=s3://lakehouse-data?host=http://${MINIO_CLUSTER_IP}:9000&s3_region=us-east-1" \
  --set lakehouse.sharedStorage.existingSecretName=existing-shared-storage-s3
```

## Immutable fields

After the custom resources are created, `deploymentMode`, `logServiceRef`,
`sharedStorageInfo`, `objectStoreUrl`, the Observer image, and the OBLogService
template cannot be changed. Changing buckets, credentials, or images normally
requires a new release and a data-migration plan. When upgrading a chart
version for an existing Lakehouse release, pass the exact Observer image that
the cluster was created with.

## Chart validation

With Helm 3, Python 3, PyYAML and jsonschema installed, run from the repository
root:

```bash
helm lint charts/oceanbase-cluster
python3 charts/oceanbase-cluster/tests/test_render.py
```

The tests cover traditional and Lakehouse rendering, generated and existing
Secrets, input validation and compatibility with the repository's CRD schemas.
They do not start a Kubernetes cluster or an OceanBase database.
