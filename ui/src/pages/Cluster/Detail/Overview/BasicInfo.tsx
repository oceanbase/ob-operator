import { COLOR_MAP,MODE_MAP } from '@/constants';
import { intl } from '@/utils/intl';
import { Card,Col,Descriptions,Switch,Tag,Typography } from 'antd';
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
  parameters,
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
          value: resource.memory,
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
          value: storage.dataStorage.size,
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
          value: storage.redoLogStorage.size,
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
          value: storage.sysLogStorage.size,
        },
      ]
    : [];

  return (
    <Col span={24}>
      <Card style={style}>
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
            <Tag color={COLOR_MAP.get(status)}>
              {status === 'operating' ? (
                <Text
                  style={{ maxWidth: 120, color: '#d48806', fontSize: 12 }}
                  ellipsis={{ tooltip: `${status}/${statusDetail}` }}
                >
                  {status}/{statusDetail}
                </Text>
              ) : (
                status
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
            {parameters && (
              <Descriptions
                title={intl.formatMessage({
                  id: 'Dashboard.Detail.Overview.BasicInfo.ClusterParameters',
                  defaultMessage: '集群参数',
                })}
              >
                {parameters.map((parameter, index) => (
                  <Descriptions.Item label={parameter.key} key={index}>
                    {parameter.value}
                  </Descriptions.Item>
                ))}
              </Descriptions>
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
          </div>
        )}
      </Card>
    </Col>
  );
}
