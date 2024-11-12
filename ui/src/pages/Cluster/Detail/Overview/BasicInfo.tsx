import { MODE_MAP, STATUS_LIST } from '@/constants';
import { floorToTwoDecimalPlaces } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { findByValue } from '@oceanbase/util';
import { Card, Descriptions, Switch, Tag, Typography } from 'antd';
import { useState } from 'react';
import styles from './index.less';

const { Text } = Typography;
export default function BasicInfo({
  name,
  namespace,
  status,
  statusDetail,
  image,
  mode,
  rootPasswordSecret,
  resource,
  storage,
  backupVolume,
  monitor,
  clusterName,
  extra = true,
  style,
}: API.ClusterInfo & { style?: React.CSSProperties; extra?: boolean }) {
  const [checked, setChecked] = useState<boolean>(false);
  const OBServerConfig = extra
    ? [
        {
          label: 'CPU',
          value: resource.cpu,
        },
        {
          label: 'Memory',
          value: floorToTwoDecimalPlaces(resource.memory / (1 << 30)) + 'Gi',
        },
        {
          label: intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.DatafileStorageClass',
            defaultMessage: 'Datafile 存储类',
          }),
          value: storage.dataStorage.storageClass,
        },
        {
          label: intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.DatafileStorageSize',
            defaultMessage: 'Datafile 存储大小',
          }),
          value:
            floorToTwoDecimalPlaces(storage.dataStorage.size / (1 << 30)) +
            'Gi',
        },
        {
          label: intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.RedologStorageClass',
            defaultMessage: 'RedoLog 存储类',
          }),
          value: storage.redoLogStorage.storageClass,
        },
        {
          label: intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.RedologSize',
            defaultMessage: 'RedoLog 大小',
          }),
          value:
            floorToTwoDecimalPlaces(storage.redoLogStorage.size / (1 << 30)) +
            'Gi',
        },
        {
          label: intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.SystemLogStorageClass',
            defaultMessage: '系统日志存储类',
          }),
          value: storage.sysLogStorage.storageClass,
        },
        {
          label: intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.SystemLogStorageSize',
            defaultMessage: '系统日志存储大小',
          }),
          value:
            floorToTwoDecimalPlaces(storage.sysLogStorage.size / (1 << 30)) +
            'Gi',
        },
      ]
    : [];

  const statusItem = findByValue(STATUS_LIST, status);
  const statusDetailItem = findByValue(STATUS_LIST, statusDetail);

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
          span={2}
          label={intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.BasicInfo.Image',
            defaultMessage: '镜像',
          })}
        >
          {image}
        </Descriptions.Item>

        <Descriptions.Item
          span={2}
          label={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.RootPasswordSecret',
            defaultMessage: 'root 用户密码 Secret',
          })}
        >
          {rootPasswordSecret || '-'}
        </Descriptions.Item>
      </Descriptions>
      {extra && (
        <div style={{ marginBottom: 12 }}>
          <span
            style={{
              color: '#132039',
              fontSize: 16,
              fontWeight: 600,
              marginRight: 8,
            }}
          >
            {intl.formatMessage({
              id: 'Dashboard.Detail.Overview.BasicInfo.DetailedClusterConfiguration',
              defaultMessage: '集群详细配置',
            })}
          </span>
          <Switch
            checked={checked}
            onChange={(checked) => setChecked(checked)}
          />
        </div>
      )}

      {checked && extra && (
        <div className={styles.detailConfig}>
          <Descriptions
            style={{ width: '50%' }}
            title={intl.formatMessage({
              id: 'Dashboard.Detail.Overview.BasicInfo.ObserverResourceConfiguration',
              defaultMessage: 'OBServer 资源配置',
            })}
          >
            {OBServerConfig.map((item, index) => (
              <Descriptions.Item span={2} key={index} label={item.label}>
                {item.value}
              </Descriptions.Item>
            ))}
          </Descriptions>

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
        </div>
      )}
    </Card>
  );
}
