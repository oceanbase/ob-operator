# Administrator preparation for Lakehouse

Prepare shared infrastructure once per cluster using the administrator's
existing kubeconfig. Install a maintained cert-manager version that supports
your Kubernetes version before running this script. For the development
acceptance target, use Kubernetes 1.35 and cert-manager 1.21. The upstream
compatibility table is at https://cert-manager.io/docs/releases/.

For the online development release, use the versioned `prepare-cluster.sh`
from the download page on a new cluster. It prepares the locked cert-manager
and shared Operator using the administrator's current kubectl context.
Use `--verify-only` to check an existing Operator.

For manual or offline preparation, extract the pinned application Chart and
run the following with Python 3 and kubectl available:

```bash
sh oceanbase-lakehouse/files/prerequisites/prepare.sh --context CONTEXT
```

This bundle contains the matched Operator payload, its OceanBase CRDs, and
image-lock.json for the two Operator containers. Mirror or preload those
images on the eligible workers before an offline install. cert-manager is
installed and updated independently by the cluster administrator.

The script creates missing objects using the invoking administrator's
credentials. It tags objects it creates with the payload digest and can
resume an interrupted run using the exact same payload. It never applies
updates to live objects. For an external Operator use `--verify-only`. The
script checks the required APIs, Deployment and admission webhooks. It
reports missing or incompatible resources for the owner to repair.

The application Chart does not install, patch or remove shared infrastructure.
Legacy prerequisites.* and skip-operator values remain accepted for upgrades.
Old releases may have an abandoned preparation ClusterRoleBinding named
<namespace>-<release>-prerequisites. After stopping legacy Helm operations
and confirming the namespace and release, the administrator can revoke it:

```bash
sh oceanbase-lakehouse/files/prerequisites/prepare.sh --context CONTEXT \
  --revoke-legacy-binding NAMESPACE RELEASE
```

The command requires the exact legacy Hook annotation, cluster-admin role
reference and sole ServiceAccount subject. It refuses active preparation Jobs
and Pods. Deletion is conditional on the binding UID and resourceVersion it
inspected. An absent binding is already clean. Helm uninstall does not track
that old Hook binding.
