# Changelog

## 2.2.0 

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
