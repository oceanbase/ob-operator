扩容和缩容概述：
[https://www.oceanbase.com/docs/common-oceanbase-database-cn-10000000001702202](https://www.oceanbase.com/docs/common-oceanbase-database-cn-10000000001702202)
## 前提
ob-operator V1.1.0 以上
ob 部署完成且正常运行
## 参考配置
[https://github.com/oceanbase/ob-operator/blob/master/deploy/obcluster.yaml](https://github.com/oceanbase/ob-operator/blob/master/deploy/obcluster.yaml)
Kubernetes 用 spec 来描述所期望的对象应该具有的状态，而用 status 字段来记录对象在系统上的当前状态。
```yaml
apiVersion: cloud.oceanbase.com/v1
kind: OBCluster
metadata:
  name: ob-test
  namespace: obcluster
spec:
  imageRepo: oceanbasedev/oceanbase-cn
  tag: v4.1.0.0-100000192023032010
  imageObagent: oceanbase/obagent:1.2.0
  clusterID: 1
  topology:
    - cluster: cn
      zone:
      - name: zone1
        region: region1
        nodeSelector:
          ob.zone: zone1
        replicas: 1
      - name: zone2
        region: region1
        nodeSelector:
          ob.zone: zone2
        replicas: 1
      - name: zone3
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1
      parameters:
        - name: log_disk_size
          value: "40G"
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
      - name: obagent-conf-file
        storageClassName: "local-path"
        size: 1Gi
    volume:
        name: backup
        nfs:
          server: ${nfs_server_address}
          path: /opt/nfs
          readOnly: false
```
上述为一个三副本的 OceanBase 集群的配置文件 obcluster.yaml 示例。当前集群中共包含 3 个可用区 zone1、zone2、zone3，每个 Zone 内包含 1 个 OBServer 节点。
## 一、扩容
### 1、向 Zone 内添加 OBServer 节点
本节主要介绍如何通过向 Zone 内添加 OBServer 节点的方式进行集群的扩容。
向 Zone 内添加 OBServer 节点的操作可以通过修改 OB 集群配置文件 obcluster.yaml 来完成。
上述配置文件 obcluster.yaml 中 spec.topology.cluster[index1].zone[index2].replicas 属性表示 在 k8s 集群 cluster[index1] 上部署的 OB 集群 中 zone[index2] 所包含的 observer 数量。通过修改该属性值可以实现向 Zone 内添加 OBServer 节点。
假如当前集群中共包含 3 个可用区 zone1、zone2、zone3，每个 Zone 内包含 1 个 OBServer 节点。现在希望向每个 Zone 内添加 1 台 OBServer 节点来扩容，则扩容后，3 个 Zone 内均包含 2 台 OBServer 节点。
```yaml
# 示例: 每个 zone 中有 1 个 observer
spec:
  topology:
    - cluster: cn
      zone:
      - name: zone1
        region: region1
        nodeSelector:
          ob.zone: zone1
        replicas: 1
      - name: zone2
        region: region1
        nodeSelector:
          ob.zone: zone2
        replicas: 1
      - name: zone3
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1
        
# 向每个 zone 内添加 1 个 observer
spec:
  topology:
    - cluster: cn
      zone:
      - name: zone1
        region: region1
        nodeSelector:
          ob.zone: zone1
        replicas: 2 # 1 -> 2
      - name: zone2
        region: region1
        nodeSelector:
          ob.zone: zone2
        replicas: 2 # 1 -> 2
      - name: zone3
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 2 # 1 -> 2
```
可以同时修改多个 zone 的 observer 数量。配置文件修改后，您需运行如下命令使改动生效。
```yaml
kubectl apply -f obcluster.yaml
```
### 2、在集群中增加 Zone
本节主要介绍如何通过在集群中增加 Zone 的方式进行集群的扩容。
在集群中增加 Zone 的操作可以通过修改 OB 集群配置文件 obcluster.yaml 来完成。
上述配置文件 obcluster.yaml 中 spec.topology.cluster[index1].zone 属性表示 在 k8s 集群 cluster[index1] 上部署的 OB 集群中 zone 的分布情况。通过修改该属性值可以实现在集群中增加 Zone 。
假如当前集群中共包含 3 个可用区 zone1、zone2、zone3，且三个 Zone 都属于同一个 Region，每个 Zone 内包含 1 个 OBServer 节点。现在希望将 3 个可用区扩容成为 5 个可用区。
```yaml
# 示例: OB 集群中有 3 个 zone
spec:
  topology:
    - cluster: cn
      zone:
      - name: zone1
        region: region1
        nodeSelector:
          ob.zone: zone1
        replicas: 1
      - name: zone2
        region: region1
        nodeSelector:
          ob.zone: zone2
        replicas: 1
      - name: zone3
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1
        
        
# 在集群中增加 zone(zone4, zone5)
spec:
  topology:
    - cluster: cn
      zone:
      - name: zone1
        region: region1
        nodeSelector:
          ob.zone: zone1
        replicas: 1
      - name: zone2
        region: region1
        nodeSelector:
          ob.zone: zone2
        replicas: 1
      - name: zone3
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1
      - name: zone4  # 增加的 zone
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1
      - name: zone5
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1
```
配置文件修改后，您需运行如下命令使改动生效。
```yaml
kubectl apply -f obcluster.yaml
```
## 二、缩容
### 1、从 Zone 中减少 OBServer 节点
本节主要介绍如何通过从 Zone 中减少 OBServer 节点的方式进行集群的缩容。
在进行集群的缩容操作前，需要确认集群中资源对当前负载有较多冗余，
从 Zone 中减少 OBServer 节点的操作可以通过修改 OB 集群配置文件 obcluster.yaml 来完成。通常需要从每个 Zone 内减少相等数量的 OBServer 节点。
上述配置文件 obcluster.yaml 中 spec.topology.cluster[index1].zone[index2].replicas 属性表示 在 k8s 集群 cluster[index1] 上部署的 OB 集群 中 zone[index2] 所包含的 observer 数量。通过修改该属性值可以实现从 Zone 中减少 OBServer 节点。
假如当前集群中共包含 3 个可用区 zone1、zone2、zone3，每个 Zone 内包含 2 个 OBServer 节点。现在希望在每个 Zone 内减少 1 台 OBServer 节点来缩容，则缩容后，3 个 Zone 内只包含 1 台 OBServer 节点。
```yaml
# 示例: 每个 zone 中有 2 个 observer
spec:
  topology:
    - cluster: cn
      zone:
      - name: zone1
        region: region1
        nodeSelector:
          ob.zone: zone1
        replicas: 2
      - name: zone2
        region: region1
        nodeSelector:
          ob.zone: zone2
        replicas: 2
      - name: zone3
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 2
        
# 每个 zone 内减少 1 个 observer
spec:
  topology:
    - cluster: cn
      zone:
      - name: zone1
        region: region1
        nodeSelector:
          ob.zone: zone1
        replicas: 1 # 2 -> 1
      - name: zone2
        region: region1
        nodeSelector:
          ob.zone: zone2
        replicas: 1 # 2 -> 1
      - name: zone3
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1 # 2 -> 1
```
配置文件修改后，您需运行如下命令使改动生效。
```yaml
kubectl apply -f obcluster.yaml
```
### 2、从集群中减少 Zone
本节主要介绍如何通过在集群中增减少 Zone 的方式进行集群的缩容。
本方案仅适用于 Zone 数量大于 3 的场景，如果集群中 Zone 的数量小于 3，则会出现不满足多数派的情况，数据一致性无法得到保证。在进行集群的缩容操作前，需要确认集群中资源对当前负载有较多冗余。
从集群中减少 Zone 的操作可以通过修改 OB 集群配置文件 obcluster.yaml 来完成。
上述配置文件 obcluster.yaml 中 spec.topology.cluster[index1].zone 属性表示 在 k8s 集群 cluster[index1] 上部署的 OB 集群中 zone 的分布情况。通过修改该属性值可以实现在集群中减少 Zone 。
假如当前集群中共包含 5 个可用区 zone1、zone2、zone3、zone4、zone5，且五个 Zone 都属于同一个 Region，每个 Zone 内包含 1 个 OBServer 节点。现在希望将 5 个可用区缩容成为 3 个可用区。
```yaml
# 示例: OB 集群中有 5 个 zone
spec:
  topology:
    - cluster: cn
      zone:
      - name: zone1
        region: region1
        nodeSelector:
          ob.zone: zone1
        replicas: 1
      - name: zone2
        region: region1
        nodeSelector:
          ob.zone: zone2
        replicas: 1
      - name: zone3
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1
      - name: zone4 
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1
      - name: zone5
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1
        
        
# 在集群中减少 zone(zone4, zone5)
spec:
  topology:
    - cluster: cn
      zone:
      - name: zone1
        region: region1
        nodeSelector:
          ob.zone: zone1
        replicas: 1
      - name: zone2
        region: region1
        nodeSelector:
          ob.zone: zone2
        replicas: 1
      - name: zone3
        region: region1
        nodeSelector:
          ob.zone: zone3
        replicas: 1
```
配置文件修改后，您需运行如下命令使改动生效。
```yaml
kubectl apply -f obcluster.yaml
```
## 

