import { obcluster } from '@/api';
import {
  CommonAffinitySpec,
  CommonAffinityType,
  CommonTolerationSpec,
  ResponseOBCluster,
} from '@/api/generated';
import { MODE_MAP, STATUS_LIST } from '@/constants';
import { intl } from '@/utils/intl';
import { findByValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import {
  Card,
  Checkbox,
  Descriptions,
  Table,
  Tag,
  Typography,
  message,
} from 'antd';
import { ColumnsType } from 'antd/es/table';
import { useMemo } from 'react';

const { Text, Title } = Typography;

interface ExtendedAffinity extends CommonAffinitySpec {
  zones: string;
}

interface ExtendedToleration extends CommonTolerationSpec {
  zones: string;
}

const nodeSelectorColumns: ColumnsType<ExtendedAffinity> = [
  {
    title: 'Zones',
    dataIndex: 'zones',
  },
  {
    title: 'Key',
    dataIndex: 'key',
  },
  {
    title: 'Operator',
    dataIndex: 'operator',
  },
  {
    title: 'Values',
    dataIndex: 'values',
    render(value: string[] | undefined) {
      return value?.join(', ') || '-';
    },
  },
  {
    title: 'Weight',
    dataIndex: 'weight',
    render(value: number | undefined) {
      return value || '-';
    },
  },
];

const affinityColumns: ColumnsType<ExtendedAffinity> = [
  {
    title: 'Type',
    dataIndex: 'type',
  },
  ...nodeSelectorColumns,
];

const tolerationColumns: ColumnsType<ExtendedToleration> = [
  {
    title: 'Zones',
    dataIndex: 'zones',
  },
  {
    title: 'Key',
    dataIndex: 'key',
  },
  {
    title: 'Operator',
    dataIndex: 'operator',
  },
  {
    title: 'Value',
    dataIndex: 'value',
    render(value: string | undefined) {
      return value || '-';
    },
  },
  {
    title: 'Effect',
    dataIndex: 'effect',
  },
  {
    title: 'TolerationSeconds',
    dataIndex: 'tolerationSeconds',
    render(value) {
      return value || '-';
    },
  },
];

interface ITopologyRendering {
  show: boolean;
  nodeSelectors: ExtendedAffinity[];
  affinities: ExtendedAffinity[];
  tolerations: ExtendedToleration[];
}

export default function BasicInfo({
  name,
  namespace,
  status,
  statusDetail,
  image,
  mode,
  rootPasswordSecret,
  backupVolume,
  monitor,
  clusterName,
  style,
  deletionProtection,
  clusterDetailRefresh,
  ...props
}: ResponseOBCluster & {
  style?: React.CSSProperties;
  extra?: boolean;
  clusterDetailRefresh?: () => void;
}) {
  const statusItem = findByValue(STATUS_LIST, status);
  const statusDetailItem = findByValue(STATUS_LIST, statusDetail);
  const topologyRendering: ITopologyRendering = useMemo(() => {
    const rendering: ITopologyRendering = {
      show: false,
      nodeSelectors: [],
      affinities: [],
      tolerations: [],
    };
    if (!props.topology) {
      return rendering;
    }
    props.topology.forEach((zone) => {
      if (zone.affinities) {
        zone.affinities.forEach((affinity) => {
          if (affinity.type === CommonAffinityType.NodeAffinityType) {
            rendering.nodeSelectors.push({
              ...affinity,
              zones: zone.name,
            });
          } else {
            rendering.affinities.push({
              ...affinity,
              zones: zone.name,
            });
          }
        });
      }
      if (zone.tolerations) {
        zone.tolerations.forEach((toleration) => {
          rendering.tolerations.push({
            ...toleration,
            zones: zone.name,
          });
        });
      }
    });
    rendering.show =
      rendering.nodeSelectors.length > 0 ||
      rendering.affinities.length > 0 ||
      rendering.tolerations.length > 0;
    return rendering;
  }, [props.topology]);

  const { runAsync: patchOBCluster, loading } = useRequest(
    obcluster.patchOBCluster,
    {
      manual: true,
      onSuccess: (res) => {
        if (res.successful) {
          message.success(
            intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.02AE8EA0',
              defaultMessage: '修改删除保护已成功',
            }),
          );
          clusterDetailRefresh?.();
        }
      },
    },
  );

  return (
    <Card
      style={style}
      title={
        <h2 style={{ marginBottom: 0 }}>
          {intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.ClusterInformation',
            defaultMessage: '集群信息',
          })}
        </h2>
      }
    >
      <Descriptions
        column={5}
        title={intl.formatMessage({
          id: 'Dashboard.Detail.Overview.BasicInfo.BasicClusterInformation',
          defaultMessage: '集群基本信息',
        })}
      >
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.BasicInfo.ClusterName',
            defaultMessage: '集群名',
          })}
        >
          {clusterName}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.ResourceName',
            defaultMessage: '资源名',
          })}
        >
          {name}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.BasicInfo.Namespace',
            defaultMessage: '命名空间',
          })}
        >
          {namespace}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.ClusterMode',
            defaultMessage: '集群模式',
          })}
        >
          {MODE_MAP.get(mode)?.text || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.BasicInfo.ClusterStatus',
            defaultMessage: '集群状态',
          })}
        >
          <Tag color={statusItem.badgeStatus}>
            {statusItem === 'operating' ? (
              <Text
                style={{ maxWidth: 110, color: '#d48806', fontSize: 12 }}
                ellipsis={{ tooltip: `${status}/${statusDetail}` }}
              >
                {statusItem.label}/{statusDetailItem.label}
              </Text>
            ) : (
              statusItem.label
            )}
          </Tag>
        </Descriptions.Item>

        <Descriptions.Item
          label={intl.formatMessage({
            id: 'src.pages.Cluster.Detail.Overview.8DB38279',
            defaultMessage: '删除保护',
          })}
        >
          <Checkbox
            // loading 态禁止操作，防止重复操作
            disabled={loading || status !== 'running'}
            defaultChecked={deletionProtection}
            onChange={(e) => {
              const body = {} as API.ParamPatchOBClusterParam;
              if (!e.target.checked) {
                body.removeDeletionProtection = true;
              } else {
                body.addDeletionProtection = true;
              }
              patchOBCluster(namespace, name, body);
            }}
          />
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'Dashboard.pages.Cluster.ClusterList.NumberOfZones',
            defaultMessage: 'Zone 数量',
          })}
        >
          {(
            props.topology?.map(
              (zone) => (zone.observers || zone.children)?.length ?? ' / ',
            ) ?? []
          ).join('-')}
        </Descriptions.Item>
        <Descriptions.Item
          span={3}
          label={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.RootPasswordSecret',
            defaultMessage: 'root 用户密码 Secret',
          })}
        >
          {rootPasswordSecret || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          span={5}
          label={intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.BasicInfo.Image',
            defaultMessage: '镜像',
          })}
        >
          {image}
        </Descriptions.Item>
      </Descriptions>
      {topologyRendering.show && (
        <div style={{ marginTop: 8, marginBottom: 32 }}>
          <Title level={5}>
            {intl.formatMessage({
              id: 'dashboard.Cluster.New.Topo.Topology',
              defaultMessage: '拓扑',
            })}
          </Title>
          {topologyRendering.nodeSelectors.length > 0 && (
            <>
              <Text style={{ marginBottom: 8, marginTop: 8, display: 'block' }}>
                Node Selector
              </Text>
              <Table
                columns={nodeSelectorColumns}
                rowKey={'zones'}
                dataSource={topologyRendering.nodeSelectors}
                pagination={false}
                size="small"
                locale={{ emptyText: <span>/</span> }}
              />
            </>
          )}
          {topologyRendering.affinities.length > 0 && (
            <>
              <Text style={{ marginBottom: 8, marginTop: 8, display: 'block' }}>
                Pod Affinity
              </Text>
              <Table
                columns={affinityColumns}
                rowKey={'zones'}
                dataSource={topologyRendering.affinities}
                pagination={false}
                size="small"
                locale={{ emptyText: <span>/</span> }}
              />
            </>
          )}
          {topologyRendering.tolerations.length > 0 && (
            <>
              <Text style={{ marginBottom: 8, marginTop: 8, display: 'block' }}>
                Tolerations
              </Text>
              <Table
                columns={tolerationColumns}
                rowKey={'zones'}
                dataSource={topologyRendering.tolerations}
                pagination={false}
                size="small"
                locale={{ emptyText: <span>/</span> }}
              />
            </>
          )}
        </div>
      )}

      {monitor && (
        <Descriptions
          title={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.MonitoringConfiguration',
            defaultMessage: '监控配置',
          })}
        >
          <Descriptions.Item label="CPU">
            {monitor.resource.cpu}
          </Descriptions.Item>
          <Descriptions.Item label="Memory">
            {monitor.resource.memory}
          </Descriptions.Item>
          <Descriptions.Item
            label={intl.formatMessage({
              id: 'Dashboard.Detail.Overview.BasicInfo.Image',
              defaultMessage: '镜像',
            })}
          >
            {monitor.image}
          </Descriptions.Item>
        </Descriptions>
      )}

      {backupVolume && (
        <Descriptions
          title={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.BackupVolumeConfiguration',
            defaultMessage: '备份卷配置',
          })}
        >
          <Descriptions.Item
            label={intl.formatMessage({
              id: 'Dashboard.Detail.Overview.BasicInfo.Address',
              defaultMessage: '地址',
            })}
          >
            {backupVolume.address}
          </Descriptions.Item>
          <Descriptions.Item
            label={intl.formatMessage({
              id: 'Dashboard.Detail.Overview.BasicInfo.Path',
              defaultMessage: '路径',
            })}
          >
            {backupVolume.path}
          </Descriptions.Item>
        </Descriptions>
      )}
    </Card>
  );
}
