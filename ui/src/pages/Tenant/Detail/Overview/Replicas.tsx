import CollapsibleCard from '@/components/CollapsibleCard';
import showDeleteConfirm from '@/components/customModal/DeleteModal';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { deleteObtenantPool } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { Button,Col,Descriptions,message } from 'antd';
import type { ClusterNSName } from '.';
import styles from './index.less';

export default function Replicas({
  replicaList,
  refreshTenant
}: {
  replicaList: API.ReplicaDetailType[];
  refreshTenant:()=>void;
}) {
  const LABEL_TEXT_MAP = {
    priority: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.Replicas.Priority',
      defaultMessage: '优先级',
    }),
    type: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.Replicas.ReplicaType',
      defaultMessage: '副本类型',
    }),
    maxCPU: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.Replicas.MaximumAvailableCpu',
      defaultMessage: '最大可用 CPU',
    }),
    memorySize: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.Replicas.MemorySize',
      defaultMessage: '内存大小',
    }),
    minCPU: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.Replicas.MinimumAvailableCpu',
      defaultMessage: '最小可用 CPU',
    }),
    logDiskSize: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.Replicas.ClogDiskSize',
      defaultMessage: 'Clog 盘大小',
    }),
  };
  const sortKeys = (keys: string[]) => {
    const minCpuIdx = keys.findIndex((key) => key === 'minCPU');
    const memorySizeIdx = keys.findIndex((key) => key === 'memorySize');
    const temp = keys[minCpuIdx];
    keys[minCpuIdx] = keys[memorySizeIdx];
    keys[memorySizeIdx] = temp;
    return keys;
  };

  const deleteZone = async (zoneName: string) => {
    const [ns, name] = getNSName();
    const res = await deleteObtenantPool({ ns, name, zoneName });
    if (res.successful) {
      refreshTenant();
      message.success(res.message || '删除成功');
    }
  };
  
  return (
    <Col span={24}>
      <CollapsibleCard
        title={
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'Dashboard.Detail.Overview.Replicas.ResourcePool',
              defaultMessage: '资源池',
            })}
          </h2>
        }
        collapsible={true}
        defaultExpand={true}
      >
        {replicaList.map((replica, index) => (
          <Descriptions
            column={5}
            key={index}
            title={
              <div className={styles.titleContainer}>
                <span>
                  {intl.formatMessage(
                    {
                      id: 'Dashboard.Detail.Overview.Replicas.ResourcePoolReplicazone',
                      defaultMessage: '资源池 - {{replicaZone}}',
                    },
                    { replicaZone: replica.zone },
                  )}
                </span>
                <div>
                  <Button type="link">编辑</Button>
                  <Button
                    onClick={() => {
                      showDeleteConfirm({
                        onOk: () => deleteZone(replica.zone),
                        title: `确定要删除该租户在${replica.zone}上的资源池吗？`,
                      });
                    }}
                    disabled={replicaList.length === 2}
                    type="link"
                    danger
                  >
                    删除
                  </Button>
                </div>
              </div>
            }
          >
            {sortKeys(Object.keys(replica)).map((key, idx) => (
              <Descriptions.Item label={LABEL_TEXT_MAP[key] || key} key={idx}>
                {replica[key]}
              </Descriptions.Item>
            ))}
          </Descriptions>
        ))}
      </CollapsibleCard>
    </Col>
  );
}
