import CollapsibleCard from '@/components/CollapsibleCard';
import showDeleteConfirm from '@/components/customModal/DeleteModal';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { deleteObtenantPool } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { Button,Col,Descriptions,message } from 'antd';
import type { OperateType } from '.';
import styles from './index.less';

interface ReplicasProps {
  replicaList: API.ReplicaDetailType[];
  refreshTenant: () => void;
  openOperateModal: (type: API.ModalType) => void;
  setEditZone: React.Dispatch<React.SetStateAction<string>>;
  editZone: string;
  operateType: React.MutableRefObject<OperateType | undefined>;
}

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

export default function Replicas({
  replicaList,
  refreshTenant,
  openOperateModal,
  setEditZone,
  operateType,
  editZone
}: ReplicasProps) {
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
      message.success(
        res.message ||
          intl.formatMessage({
            id: 'Dashboard.Detail.Overview.Replicas.DeletedSuccessfully',
            defaultMessage: '删除成功',
          }),
      );
    }
  };

  const editResourcePool = (zone: string) => {
    operateType.current = 'edit';
    setEditZone(zone);
    openOperateModal('modifyUnitSpecification');
  };
  
  const addResourcePool = ()=>{
    operateType.current = 'create';
    openOperateModal('modifyUnitSpecification');
  }

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
        extra={<Button type='primary' onClick={addResourcePool}>新增资源池</Button>}
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
                  <Button
                    onClick={() => editResourcePool(replica.zone)}
                    type="link"
                  >
                    {intl.formatMessage({
                      id: 'Dashboard.Detail.Overview.Replicas.Edit',
                      defaultMessage: '编辑',
                    })}
                  </Button>
                  <Button
                    onClick={() => {
                      showDeleteConfirm({
                        onOk: () => deleteZone(replica.zone),
                        title: intl.formatMessage(
                          {
                            id: 'Dashboard.Detail.Overview.Replicas.AreYouSureYouWant',
                            defaultMessage:
                              '确定要删除该租户在{{replicaZone}}上的资源池吗？',
                          },
                          { replicaZone: replica.zone },
                        ),
                      });
                    }}
                    disabled={replicaList.length === 2 || replicaList.length === 1}
                    type="link"
                    danger
                  >
                    {intl.formatMessage({
                      id: 'Dashboard.Detail.Overview.Replicas.Delete',
                      defaultMessage: '删除',
                    })}
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
