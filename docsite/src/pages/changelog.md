# Changelog

## 2.3.3 (Release on 2025.09.08)
### New Feature
1. Support setting addressingModel for S3_COMPATIBLE storage.
2. Support setting variables while creating tenant, for setting variables which are not able to be modifed after tenant creation.

## 2.3.2 (Release on 2025.06.12)
### New Feature
1. Support to keep static IP address when using kube-ovn as network plugin

### Bugfix
1. Fix connection problem when upgrading single node cluster.
2. Fix monitor password config problem when creating obcluster with obagent in multiple K8s clusters.

## 2.3.1 (Release on 2024.11.28)

### New Feature
1. Support tenant variable and parameter management.

### Bugfix
1. Fix obagent deployment with service or standalone mode.
2. Fix sql execution failure when adding server to the cluster in certain scenario.

### Enhancement
1. Add image and store config to the output of obcluster resource.

## 2.3.0 (Release on 2024.10.14)

### New Features

1. Support for scheduling OceanBase cluster across multiple K8s clusters.
2. Support for backing up to Tencent COS, AWS s3 and s3 compatible storage.
3. Support for deleting specific OBServer.
4. Support for optimizing parameters and variables by scenario.
5. Support for setting most of native fields of Pods to OBServer.

### Bug fixes

1. Fixed the issue that it get stuck when a 2-2-2 cluster rolling replace its OBServer pods.

### Optimization

1. Supplement several new types of `OBTenantOperation` to perform common operations like creating or deleting resource pools, setting unit number and so on.

## 2.2.2 (Release on 2024.06.18)

### New Features

1. Support for default storage class by leaving `storageClass` fields empty. 
2. Support for pausing reconciling CRD through `/pause-reconciling: "true"` annotation. Which may become a solution of issue. 
3. Support for changing storage class of running OBClusters. 
4. New CRD `OBClusterOperation` for common operations of OBClusters. 

### Bug fixes

1. Fixed issue that cluster may get deactivated after being shrunk by half.

## 2.2.1 (Released on 2024.05.15)

### New features

1. Support for customizing configuration of ob-operator by setting environment variables.
2. Support for attaching backup volume after OceanBase cluster started if static ip is supported.
3. Support for deleting protection of important resources by adding annotation.
4. Support for migrating OceanBase cluster to be managed by ob-operator.

### Bug fixes

1. Fixed issue when creating backup tenant across namespace.
2. Fixed issue when recover observer if OceanBase cluster is not accessible.

### Optimization

1. Checked log storage volume fit OceanBase's requirement before actually create the cluster.
2. Optimized version check when creating OceanBase cluster in service mode.

## 2.2.0 (Released on 2024.03.28)

### New features

1. Support for binding static IP address of observer without depending on certain CNI by introducing the `service` mode of OBCluster.
2. Support for correcting `Status` of OceanBase related resources using `delete`, `reset`, `skip` and `retry` operations offered by CRD `OBResourceRescue`.
3. Support for expanding PVCs dynamically when the OBCluster is running.

### Bug fixes

1. Fixed infinite retry when creating a tenant that requires resource (CPU, Memory etc.) more than available.

### Optimization

1. Accelerated reconciliation when sub resource changes.
2. Exposed metrics of controller manager through prometheus standard interface.

## 2.1.2 (Released on 2024.01.24)

### New features

1. Support for creating OceanBase clusters in standalone mode by adding annotation `"oceanbase.oceanbase.com/mode": "standalone"`.
2. Support for scaling cluster's resources (CPU and Memory) in place in standalone mode.
3. Support for binding single PVC with pods by setting annotation `"oceanbase.oceanbase.com/single-pvc": "true"`.
4. Support for binding a service account with pods by configuring `spec.serviceAccount` for OBCluster.

### Bug fixes

1. Fixed unexpected behavior when ob-operator restarts during observer recovery.
2. Avoided random deletion of OBServer when modifying OBZone replicas.

### Optimization

1. Optimized task manager by adding token pool to limit maximum number of concurrent running tasks.
2. Optimized database connection pool by replacing `sync.Map` with expirable LRU cache.
3. Reduced required user credentials to create OBCluster. Automatically create credentials for user `proxyro`, `monitor` and `operator` if not specified.
4. Reduced required user credentials to create OBTenant. Automatically create credentials for user `root` and `standbyro` if not specified.
5. Ensured the deploying image is ready before the OceanBase cluster is actually created.

## 2.1.1 (Released on 2023.12.20)

### New features

1. Support for adding annotation `oceanbase.oceanbase.com/independent-pvc-lifecycle` to the `OBCluster` resource to make PVC remain after the `OBCluster` resource is deleted.
2. Support for tenant upgrade with the `OBTenantOperation` resource, which is a feature introduced since OceanBase Database V4.1.
3. Support for setting cluster parameters with `optstr` in startup command of `observer`.

### Bug fixes

1. Fixed the issue of unrestricted memory consumption in some container runtimes by explicitly setting the `memory_limit` parameter.
2. Avoided long waits for changes to take effect when parameters are altered after cluster bootstraps by setting these parameters during startup.

### Optimization

1. Reduced initial value of `datafile_size` and use incremental step to scale up when needed.
2. Enhanced resource validation, especially for `OBCluster` and `OBTenant` resources.

## 2.1.0 (Released on 2023.11.20)

### New features

1. Added toleration and affinity options to cluster parameters.
2. Support for restoring tenant and creating standby tenant from backup data.
3. The ARM image is now provided.

### Bug fixes

1. Fixed issues that may be caused by concurrent writes to map.
2. Fixed the issue where expired database connections were used during tenant backup.
3. Fixed the issue of SQL syntax compatibility with OceanBase Database 4.2.1.

### Optimization

1. Optimized failed task retry with backoff.
2. Added event logging and improved log outputs.

## 2.0.0 (Released on 2023.09.26)

### New features

1. Support OceanBase Cluster management.
2. Support Oceanbase Tenant management.
3. Support for monitoring OceanBase database with OBAgent.
