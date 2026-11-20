# Optional MinIO application chart

A single-node MinIO StatefulSet, Service, PVC and retained root credentials.
This chart does not install the MinIO Operator or manage lakehouse business users.

The lakehouse parent aliases this dependency as `minio` and uses
`condition: minio.enabled`. The parent independently selects its storage endpoint:
a nonempty `objectStorage.endpoint` always wins, even when this dependency is enabled.

All values are local to this chart. Configure `storageClass`, `region`, `nodeSelector`,
`tolerations` and `imagePullSecrets` here (under `minio.*` in the parent). Existing
resource names and pod selectors match the former parent templates. Keep the same
release/namespace and existing data configuration when upgrading. Retained resources
are not deleted or stopped when the parent disables this dependency.

Root Secrets require `rootUser` and `rootPassword`. `rootExistingSecret` and
`existingClaim` reference resources managed separately. Restore missing retained
credentials instead of generating replacements. The retained identity prevents
changing the root Secret, data claim, storage class or capacity without migration.
MinIO images are optional third-party software licensed under AGPL-3.0; images are
pulled separately and are not included in this chart.


Set `mode: distributed` to select the three-worker, one-RWO-claim-per-Pod
layout. `distributed.driveSize` sets capacity per claim. The client Service is
shared across layouts, and a separate headless Service discovers peers. The
initial topology, storage class and layout are immutable for retained data.
See the parent chart's SHARED-INFRA.md for scheduling and recovery guidance.
