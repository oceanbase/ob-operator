租户管理概述：
[https://www.oceanbase.com/docs/common-oceanbase-database-cn-10000000001702213](https://www.oceanbase.com/docs/common-oceanbase-database-cn-10000000001702213) 
OceanBase 数据库采用了单集群多租户设计，一个集群内可包含多个相互独立的租户。在 OceanBase 数据库中，租户是资源分配的单位，是数据库对象管理和资源管理的基础。
租户按照兼容模式的不同，又分为 MySQL 租户和 Oracle 租户。ob-operator 支持管理  MySQL 租户。
## 前提
ob-operator V1.2.0-snapshot 以上
ob 部署完成且正常运行

## 一、新建租户
您可以通过应用租户配置文件 tenant.yaml 新建租户。
[https://github.com/oceanbase/ob-operator/blob/master/deploy/tenant.yaml](https://github.com/oceanbase/ob-operator/blob/master/deploy/tenant.yaml)
创建租户的命令如下，该命令会在当前 k8s 集群中新建的一个租户 CR (custom resource) 。
```yaml
kubectl apply -f tenant.yaml
```
### 参考配置 tenant.yaml 
配置风格为 cluster 风格。
创建名为 obtenant 的一个 3 副本的 mysql 租户，并指定允许任何客户端 IP 连接该租户。
创建租户时，会根据配置文件 tenant.yaml 中的 topology 为其中的每一个 zone 创建一个关联的 resource unit 和 resource pool。其中创建 resource unit 所需要的的规格配置即为 resource 下的配置项；创建 resource pool 所需要的配置中 unit 为该 zone 对应的 resource unit、unit_num 为 unitNum 配置项、zone_list 即为该 zone。
```yaml
apiVersion: cloud.oceanbase.com/v1		
kind: Tenant		
metadata:		
  name: obtenant
  namespace: obcluster		
spec:	
  # 必填
  clusterName: ob-test		
  clusterID: 1
  # 非必填
  charset:  
  collate: 
  connectWhiteList: '%'

  topology:
    - zone: zone1
      unitNum: 1
      type: 
        name: FUll 
        replica: 1
      priority: 3
      resource:
        maxCPU: 2500m
        memorySize: 1Gi
        # 非必填
        minCPU: 2 
        maxIops: 1024 
        minIops: 1024 
         # v3
        maxDiskSize: 3Gi 
        maxSessionNum:  512
         # v4
        iopsWeight: 
        logDiskSize: 
    - zone: zone2
      unitNum: 1
      type: 
        name: Full
        replica: 1
      priority: 3
      resource:
        maxCPU: 2500m 
        memorySize: 1Gi
        # 非必填
        minCPU: 2 
        maxIops: 1024
        minIops: 1024
         # v3
        maxDiskSize: 1Gi
        maxSessionNum: 64 
         # v4
        iopsWeight: 2
        logDiskSize: 4Gi 
    - zone: zone3
      unitNum: 1
      type: 
        name: Full
        replica: 1
      priority: 3
      resource:
        maxCPU: 2500m 
        memorySize: 1Gi
        # 非必填
        minCPU: 2 
        maxIops: 1024
        minIops: 1024
         # v3
        maxDiskSize: 1Gi
        maxSessionNum: 64 
         # v4
        iopsWeight: 2
        logDiskSize: 4Gi 

```
配置项说明如下

| 配置项 | 说明 |
| --- | --- |
| metadata.name | 租户名的合法性和变量名一致，最长 128 个字符，字符只能是大小写英文字母、数字和下划线，而且必须以字母或下划线开头，并且不能是 OceanBase 数据库的关键字。 OceanBase 数据库中所支持的关键字请参见 MySQL 模式的 [预留关键字](https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000001687907) 和 Oracle 模式的 [预留关键字](https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000001687908)；必填 |
| metadata.namespace | 指定租户所在的命名空间；必填 |
| clusterName | 指定需要创建租户的 ob 集群名；必填 |
| clusterID | 指定需要创建租户的 ob 集群 ID；必填 |
| charset | 指定租户的字符集，字符集相关的介绍信息请参见 [字符集](https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000001702724)；非必填，默认设置为 'utf8mb4' |
| collate | 指定租户的字符序，字符序相关的介绍信息请参见 [字符序](https://www.oceanbase.com/docs/enterprise-oceanbase-database-cn-10000000001702725)；非必填 |
| connectWhiteList | 指定允许连接该租户的客户端 IP，'%' 表示任何客户端 IP 都可以连接该租户；非必填，默认设置为 '%'。如果用户需要修改改配置，则需要将 ob-operator 所处的网段包含在配置内，否则 ob-operator 会连接不上该租户。 |
| topology | 租户的拓扑结构，用于定义租户在每个 zone 上的副本、资源分布等情况。 |
| zone | 指定可用区 zone名；必填 |
| unitNum | 指定要创建的单个 ZONE 下的单元个数，取值要小于单个 ZONE 中的 OBServer 个数；必填 |
| type.name | 指定租户在该 zone 的副本类型，支持full、logonly、readonly, 需要写出完整类型, 大小写不敏感；必填 |
| type.replica | 指定租户在该 zone 的副本数；非必填，默认为 1 |
| priority | 指定当前 zone 的优先级，数字越大优先级越高；非必填，默认为 0 |
| resource | 指定租户在该 zone 的资源情况 |
| maxCPU | 指定租户在该 zone 上 使用的资源单元提供的 CPU 的上限；必填，V3、V4最小值为 1 |
| memorySize | 指定租户在该 zone 上 使用的资源单元提供的 Memory 的大小；必填，最小值为 1Gi |
| minCPU | 指定租户在该 zone 上 使用的资源单元提供的 CPU 的下限；非必填，默认等于 maxCPU |
| maxIops | 指定租户在该 zone 上 使用的资源单元提供的 Iops 的上限；非必填，V3 默认等于 128，V4 默认有计算规则 [https://www.oceanbase.com/docs/common-oceanbase-database-cn-10000000001699430](https://www.oceanbase.com/docs/common-oceanbase-database-cn-10000000001699430)， |
| minIops | 指定租户在该 zone 上 使用的资源单元提供的 Iops 的下限；非必填，V3 默认等于128，V4 默认有计算规则 [https://www.oceanbase.com/docs/common-oceanbase-database-cn-10000000001699430](https://www.oceanbase.com/docs/common-oceanbase-database-cn-10000000001699430) |
| maxDiskSize | 指定租户在该 zone 上 使用的资源单元提供的磁盘大小的上限；非必填，仅 V3 需要设置该参数，默认等于 512Mi |
| maxSessionNum | 指定租户在该 zone 上 使用的资源单元提供的 session 数的上限；非必填，仅 V3 需要设置该参数，默认等于 64 |
| iopsWeight | 指定租户在该 zone 上 使用的资源单元提供的 Iops 权重。非必填，仅 V4 需要设置该参数，默认等于 1， |
| logDiskSize | 指定租户在该 zone 上 使用的资源单元提供的日志盘规格。非必填，仅 V4 需要设置该参数，默认等于 3 倍的内存规格，最小值为 2Gi |

### 确认租户是否创建成功
创建租户后，您可以通过查询 tenants.cloud.oceanbase.com 这个 CR (custom resource) 来确认租户是否创建成功。
执行以下语句，查看当前 k8s 集群中是否有新创建的租户 CR，并且该 CR 的 Status.status 为 Running，相关配置都会在 Status 中展示。
```yaml
kubectl get tenants.cloud.oceanbase.com -A -o yaml
```
示例如下：
```yaml
apiVersion: v1
items:
- apiVersion: cloud.oceanbase.com/v1
  kind: Tenant
  metadata:
    annotations:
      kubectl.kubernetes.io/last-applied-configuration: |
        {"apiVersion":"cloud.oceanbase.com/v1","kind":"Tenant","metadata":{"annotations":{},"name":"obtenant","namespace":"obcluster"},"spec":{"charset":null,"clusterID":1,"clusterName":"ob-test","collate":null,"connectWhiteList":"%","topology":[{"priority":3,"resource":{"iopsWeight":null,"logDiskSize":null,"maxCPU":"2500m","maxDiskSize":"3Gi","maxIops":1024,"maxSessionNum":512,"memorySize":"1Gi","minCPU":2,"minIops":1024},"type":{"name":"FUll","replica":1},"unitNum":1,"zone":"zone1"},{"priority":3,"resource":{"iopsWeight":2,"logDiskSize":"4Gi","maxCPU":"2500m","maxDiskSize":"1Gi","maxIops":1024,"maxSessionNum":64,"memorySize":"1Gi","minCPU":2,"minIops":1024},"type":{"name":"Full","replica":1},"unitNum":1,"zone":"zone2"},{"priority":3,"resource":{"iopsWeight":2,"logDiskSize":"4Gi","maxCPU":"2500m","maxDiskSize":"1Gi","maxIops":1024,"maxSessionNum":64,"memorySize":"1Gi","minCPU":2,"minIops":1024},"type":{"name":"Full","replica":1},"unitNum":1,"zone":"zone3"}]}}
    creationTimestamp: "2023-04-03T12:38:15Z"
    finalizers:
    - cloud.oceanbase.com.finalizers.obtenant
    generation: 2
    name: obtenant
    namespace: obcluster
    resourceVersion: "3348758"
    uid: 9710ab83-3457-4552-88aa-d25a26f2898d
  spec:
    clusterID: 1
    clusterName: ob-test
    connectWhiteList: '%'
    ......
  status:
    connectWhiteList: '%'
    status: Running
    ......
kind: List
metadata:
  resourceVersion: ""
```
## 二、修改租户
您可以通过应用租户配置文件 tenant.yaml 修改租户。
[https://github.com/oceanbase/ob-operator/blob/master/deploy/tenant.yaml](https://github.com/oceanbase/ob-operator/blob/master/deploy/tenant.yaml)
### 1、修改资源规格
如果您需要修改租户的资源规格配置，可以修改 resource 下的配置项从而修改在该 zone 下的 resource unit 的规格。
```yaml
 resource:
      maxCPU: 2500m 
      memorySize: 1Gi
      # 非必填
      minCPU: 2 
      maxIops: 1024
      minIops: 1024
       # v3
      maxDiskSize: 1Gi
      maxSessionNum: 64 
       # v4
      iopsWeight: 2
      logDiskSize: 4Gi 
```
配置文件修改后，您需运行如下命令使改动生效。
```yaml
kubectl apply -f tenant.yaml
```
可以通过查看租户 CR 的 Status 中的 resource 来判断修改是否成功。
执行以下语句，查看当前 k8s 集群中的租户 CR。
```yaml
kubectl get tenants.cloud.oceanbase.com -A -o yaml
```
```yaml
status:
    ......
    topology:
    - zone: zone1
      resource:
        logDiskSize: "3221225472"
        maxCPU: 2500m
        maxDiskSize: "0"
        maxIops: 1024
        memorySize: "1073741824"
        minCPU: "2"
        minIops: 1024
```
### 2、修改租户 Primary Zone
如果您需要修改租户的 Primary Zone，可以修改 priority 配置项用于指定 zone 的优先级，数字越大优先级越高。最小值为 1。
imary Zone 描述了 Leader 副本的偏好位置，而 Leader 副本承载了业务的强一致读写流量，即 Primary Zone 决定了 OceanBase 数据库的流量分布。通过修改 Primary Zone 属性可以切换业务流量，或者是从一个机房切换到另一个机房，或者是从一个城市切换到另一个城市，适用于容灾场景、扩缩容等场景。
```yaml
- zone: zone1
  priority: 1
- zone: zone2
  priority: 2
- zone: zone3
  priority: 3
  
 # 修改后
- zone: zone1
  priority: 3
- zone: zone2
  priority: 2
- zone: zone3
  priority: 1
```
例如上述配置，租户原本的 primary zone 为 'zone3;zone2;zone1'，修改优先级后租户的 primary zone 为 'zone1;zone2;zone3'
可以通过查看租户 CR 的 Status 中的 priority 来判断修改是否成功。
执行以下语句，查看当前 k8s 集群中的租户 CR。
```yaml
kubectl get tenants.cloud.oceanbase.com -A -o yaml
```
```yaml
status:
    ......
    topology:
    - zone: zone1
      priority: 3
    - zone: zone2
      priority: 2
    - zone: zone3
      priority: 1
```
在status中，priority的数字可能和spec的数字不一样，但是只要表示的优先级顺序一样即可。
如zone1:1, zone2:1, zone3:2 和 zone1:1, zone2:1, zone3:3 表示的优先级顺序一样，都是 zone3; zone1, zone2;
### 3、修改租户 locality
如果您需要修改租户的 locality，可以修改 type 配置项用于指定副本类型和副本数量，副本类型支持full、logonly、readonly。
```yaml
 type: 
    name: Full
    replica: 1
```
可以通过查看租户 CR 的 Status 中的 type 来判断修改是否成功。
### 4、修改资源池 unit num
如果您需要修改租户在某个zone的资源池属性 unit num，可以修改该 zone 下的 unitNum 配置项。
```yaml
- zone: zone3
    unitNum: 1
```
可以通过查看租户 CR 的 Status 中的 unit num 来判断修改是否成功。
### 5、修改租户的连接白名单
如果您需要修改租户的客户端 IP 连接白名单，可以修改该租户的 connectWhiteList 配置项。如果用户需要修改改配置，则需要将 ob-operator 所处的网段包含在配置内，否则 ob-operator 会连接不上该租户。
```yaml
spec:	
  ......
  connectWhiteList: '%'
```
可以通过查看租户 CR 的 Status 中的 connectWhiteList 来判断修改是否成功。
### 6、修改租户的 resource pool list
如果您需要修改租户的 resource pool list，可以修改该租户的 topology 中 zone 。每个zone 都对应一个resource pool，只需要增加某个 zone 的配置或者删除某个 zone 的配置就可以实现增加或者删除租户的resource pool list 中的某个zone。
```yaml
  - zone: zone2
      unitNum: 1
      type: 
        name: Full
        replica: 1
      priority: 3
      resource:
        maxCPU: 2500m 
        memorySize: 1Gi
        # 非必填
        minCPU: 2 
        maxIops: 1024
        minIops: 1024
         # v3
        maxDiskSize: 1Gi
        maxSessionNum: 64 
         # v4
        iopsWeight: 2
        logDiskSize: 4Gi 
```
可以通过查看租户 CR 的 Status 中的 topology 来判断修改是否成功。
## 三、删除租户
您可以通过应用删除租户配置文件 tenant.yaml 来删除租户。
[https://github.com/oceanbase/ob-operator/blob/master/deploy/tenant.yaml](https://github.com/oceanbase/ob-operator/blob/master/deploy/tenant.yaml)
删除租户的命令如下，该命令会在当前 k8s 集群中删除对应的租户 CR (custom resource) 。
```yaml
kubectl delete -f tenant.yaml
```
执行以下语句，查看当前 k8s 集群中是否有刚才删除的租户 CR。
```yaml
kubectl get tenants.cloud.oceanbase.com -A -o yaml
```
