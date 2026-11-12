# Install another lakehouse, sharing existing infrastructure

Independent controls:
- skip-operator=true: reuse the cluster's Operator and cert-manager. Pass
  Helm --skip-crds to skip the cluster-wide CRD installation phase. An
  existing Operator must serve the required CRDs and admission webhooks.
- minio.enabled=false: do not deploy the optional MinIO subchart.
- objectStorage.endpoint: use this external storage address regardless of minio.enabled; supply region, logBucket,
  dataBucket, and existingSecret (keys access_id/access_key in the new namespace).
- dashboard.enabled=false: do not deploy Dashboard or its monitoring; do not wait for it.
  Existing Dashboard permissions/version and namespace/monitoring discovery are admin-managed.

Edit values-existing-infra.yaml, replace S3 endpoint, buckets and Secret name.
Use dedicated buckets for each installation. Create the credential Secret securely
in the NEW namespace before installing. Never reuse an existing database's buckets.

helm upgrade --install ob-second ./oceanbase-lakehouse-*.tgz \
  -n ob-second --create-namespace --skip-crds -f values-existing-infra.yaml \
  --wait --wait-for-jobs --timeout 35m

For an additional MinIO or Dashboard, independently set minio.enabled=true or
dashboard.enabled=true. To use internal MinIO, also remove external
endpoint/secret settings. objectStorage.mode is deprecated and ignored.
For bundled MinIO, buckets can default to release-scoped names.

SQL and new Dashboard NodePorts default to 0 (automatic free-port allocation).
On upgrades the existing Service port is retained. To choose a fixed free port,
set sqlService.nodePort or dashboard.service.nodePort. Discover allocated ports:
  kubectl -n ob-second get svc

Use separate release and namespace names, bucket names and appropriate cluster IDs.
Do not use these switches to migrate storage or remove components from an existing
installation. Existing values and storage identities must be kept on upgrades.
Uninstall does not remove shared Operator, external S3 or a separately managed Dashboard.
Retained database CRs/PVCs/secrets follow the package retention policy; it is not a data wipe.
See build-validation.json for the delivery's tested environments, results and limitations.
Functional acceptance with cached images does not establish cold-download performance.

On a fresh cluster, the published Chart registers CRDs before rendering the
lakehouse resources. Its pre-install hooks create missing shared components and
wait for admission webhooks. Shared resources persist after application uninstall.
SS-specific monitoring supports shared namespace discovery. Traditional OBAgent
monitoring needs separate configuration.

## Slow initial downloads

A three-node cold test hit the default 30-minute readiness deadline while one node
was still downloading the Observer image. Warm-cache functional results do not
mean that this cold attempt passed. Prepare the exact offline images on every
eligible node, or explicitly extend the final readiness budget:

  helm upgrade --install ob-demo ./oceanbase-lakehouse-*.tgz \
    -n ob-demo --create-namespace -f values-slow-network.yaml \
    --wait --wait-for-jobs --timeout 70m

The profile sets native.readiness.timeoutSeconds=3600. It does not change the
storage initialization deadline. Helm timeout applies to its
individual stages; an external driver must allow their cumulative duration.
If an interrupted CLI leaves pending-install, do not edit Helm Secrets or delete
the namespace. Inspect initialization and preserve evidence first. Recovery on
a test instance can use controlled uninstall --keep-history and same-name,
same-namespace reinstall after checking retained state. Never assume uninstall
is a general-purpose recovery procedure for arbitrary production data.

## Storage selection and optional MinIO deployment

| External endpoint | minio.enabled | Result |
| --- | --- | --- |
| empty | false | Configuration error |
| empty | true | Deploy and use internal MinIO |
| configured | false | Use external storage only |
| configured | true | Deploy MinIO, use external storage |

The MinIO application subchart supports a standalone instance and the opt-in
three-member layout described below.
When external storage is selected, initialization uses only the external credentials;
it does not provision buckets or business users on the unused MinIO deployment.
Set minio.storageClass, minio.region, minio.nodeSelector, minio.tolerations and
minio.imagePullSecrets independently of the parent settings. When upgrading older
values with custom parent settings, copy those settings under minio before upgrading.
Internal storage requires matching minio.region and objectStorage.region.
Resource names, PVCs and credentials are preserved. Disabling the dependency stops
rendering its resources; existing resources annotated keep are intentionally retained.
It does not shut down a previously deployed server or delete its data automatically.
Changing the database's actual storage still requires an explicit migration. The first
upgrade verifies the previous storage identity; subsequent upgrades exclude unused
MinIO settings from the database identity. The subchart protects its own identity.

## Upgrade from the parent-owned MinIO templates

The installer automatically migrates effective values from installed legacy releases.
For a direct Helm upgrade, export the installed effective values and convert them
before using the new chart. Run in a protected local directory (effective values
can include user-provided passwords); the tool writes a new file with mode 0600.
Python 3 and PyYAML are required, as for the offline packaging tools.

```bash
umask 077
helm get values ob-demo -n ob-demo --all -o yaml > previous-values.yaml
python3 tools/migrate-minio-values.py previous-values.yaml --output migrated-values.yaml
helm upgrade ob-demo ./oceanbase-lakehouse-*.tgz -n ob-demo \
  -f migrated-values.yaml --wait --wait-for-jobs --timeout 35m
```

For pre-0.5.7 Dashboard upgrades, also follow the private backup and explicit
history migration procedure in USAGE.txt before enabling rollbacks.

The conversion copies legacy storageClass, region, selectors, tolerations and
imagePullSecrets into minio.*, and preserves the old deployment decision as
minio.enabled. Already migrated values and explicit MinIO overrides are preserved.
It does not change endpoint, buckets, identities, credentials or existing data.
The chart refuses to adopt legacy internal MinIO with mismatched inherited
settings, before applying changes. Do not use --reuse-values as a replacement for
this conversion. After migration, minio.enabled and endpoint are independent.


## Distributed bundled MinIO

The default MinIO layout is standalone. For a fresh native Helm release, apply
`-f values-three-nodes.yaml -f values-minio-ha.yaml`. The optional layout runs
three MinIO Pods on distinct hosts, with one RWO claim per Pod. A healthy
three-drive deployment can retain read/write quorum through one host loss.
Verify host failure and recovery in the target cluster before relying on HA.

Set `minio.storageClass` to the actual provisioner and
`minio.distributed.driveSize` to the capacity of **each** claim. For node-local
storage, use topology-aware binding (WaitForFirstConsumer) and provide capacity
on all three selected hosts. Use a common node label if restricting the pool.
A local PV stays tied to its node; replacing a failed node or disk requires
recovery of that member. The PDB allows one voluntary disruption. Restart
members one at a time and wait for full health before restarting another.

The S3 Service remains `<release>-minio:9000`. The separate headless Service
is used for peer discovery. External endpoint precedence is the same in both
layouts. The distributed mode uses per-Pod claims and cannot use
`minio.existingClaim`. Switch layouts by deploying a new release and performing
an explicit data migration. The distributed profile requires the Lakehouse
MinIO image built with `scripts/lakehouse/images/minio-rejoin/build.sh`.
If overriding `minio.image`, use an image with the same
`/minio/health/lakehouse-ready` contract. Each new process checks all three
disk connections before accepting client traffic. Previously admitted processes
continue serving when they have write quorum. If every process restarts
while a peer is unavailable, the S3 Service waits for all three peers.

Keep the existing layout and claim configuration
when upgrading, reinstalling retained resources, or rolling back a release.

## Database volume StorageClasses

Each field overrides only the matching volume. Omit it or set it to an empty
string to inherit top-level `storageClass`.

```yaml
storageClass: local-path
observer:
  dataStorageClass: fast-data
  logStorageClass: process-logs
logService:
  storeStorageClass: fast-store
  logStorageClass: process-logs
```

Create the named classes before installation. Both volumes for a Pod must be
provisionable on the same eligible node. In shared-storage mode, the Observer
log volume holds process logs. LogService store and log volumes are separate.
These settings select classes at creation. Moving existing data to another
class requires an explicit migration. Both split classes and distributed
MinIO use the native Helm entry point.
