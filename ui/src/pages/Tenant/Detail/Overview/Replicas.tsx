import CollapsibleCard from '@/components/CollapsibleCard';
import showDeleteConfirm from '@/components/customModal/DeleteModal';
import { useParams } from '@umijs/max';
import { deleteObtenantPool } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { Button, Col, Descriptions, message } from 'antd';
import styles from './index.less';

interface ReplicasProps {
  replicaList: API.ReplicaDetailType[];
  tenantStatus: string;
  refreshTenant: () => void;
  openOperateModal: (type: API.ModalType) => void;
  setEditZone: React.Dispatch<React.SetStateAction<string>>;
  cluster: API.SimpleCluster;
  operateType: React.MutableRefObject<OBTenant.OperateType | undefined>;
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
  cluster,
  tenantStatus
}: ReplicasProps) {
  const { ns, name } = useParams();
  const sortKeys = (keys: string[]) => {
    const minCpuIdx = keys.findIndex((key) => key === 'minCPU');
    const memorySizeIdx = keys.findIndex((key) => key === 'memorySize');
    const temp = keys[minCpuIdx];
    keys[minCpuIdx] = keys[memorySizeIdx];
    keys[memorySizeIdx] = temp;
    return keys;
  };

  const deleteResourcePool = async (zoneName: string) => {
    const res = await deleteObtenantPool({ ns:ns!, name:name!, zoneName });
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
    openOperateModal('editResourcePools');
  };

  const addResourcePool = () => {
    operateType.current = 'create';
    openOperateModal('createResourcePools');
  };

  return (
    <Col span={24}>
      <CollapsibleCard
        loading={!cluster?.topology?.length}
        title={
          <h2 style={{ marginBottom: 0 }}>
            {intl.formatMessage({
              id: 'Dashboard.Detail.Overview.Replicas.ResourcePool',
              defaultMessage: '资源池',
            })}
          </h2>
        }
        extra={
          <Button
            type="primary"
            disabled={cluster?.topology?.length === replicaList.length || tenantStatus !== 'running'}
            onClick={addResourcePool}
          >
            {intl.formatMessage({
              id: 'Dashboard.Detail.Overview.Replicas.AddAResourcePool',
              defaultMessage: '新增资源池',
            })}
          </Button>
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
                  <Button
                    onClick={() => editResourcePool(replica.zone)}
                    disabled={tenantStatus !== 'running'}
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
                        onOk: () => deleteResourcePool(replica.zone),
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
                    disabled={
                      replicaList.length <= 2 || tenantStatus !== 'running'
                    }
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
