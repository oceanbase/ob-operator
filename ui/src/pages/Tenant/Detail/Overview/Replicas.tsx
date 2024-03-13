import CollapsibleCard from '@/components/CollapsibleCard';
import { intl } from '@/utils/intl';
import { Col, Descriptions } from 'antd';

export default function Replicas({
  replicaList,
}: {
  replicaList: API.ReplicaDetailType[];
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
            key={index}
            column={5}
            title={intl.formatMessage(
              {
                id: 'Dashboard.Detail.Overview.Replicas.ResourcePoolReplicazone',
                defaultMessage: '资源池 - {{replicaZone}}',
              },
              { replicaZone: replica.zone },
            )}
          >
            {sortKeys(Object.keys(replica)).map((key, idx) => {
              console.log('key', key);

              return (
                <Descriptions.Item label={LABEL_TEXT_MAP[key] || key} key={idx}>
                  {replica[key]}
                </Descriptions.Item>
              );
            })}
          </Descriptions>
        ))}
      </CollapsibleCard>
    </Col>
  );
}
