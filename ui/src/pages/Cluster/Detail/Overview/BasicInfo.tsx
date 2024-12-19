import { obcluster } from '@/api';
import { MODE_MAP, STATUS_LIST } from '@/constants';
import { intl } from '@/utils/intl';
import { findByValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import { Card, Checkbox, Descriptions, Tag, Typography, message } from 'antd';

const { Text } = Typography;
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
  deletionProtection
}: API.ClusterInfo & {style?: React.CSSProperties;extra?: boolean;}) {
  const statusItem = findByValue(STATUS_LIST, status);
  const statusDetailItem = findByValue(STATUS_LIST, statusDetail);

  const { runAsync: patchOBCluster, loading } = useRequest(
    obcluster.patchOBCluster,
    {
      manual: true,
      onSuccess: (res) => {
        if (res.successful) {
          message.success(intl.formatMessage({ id: "src.pages.Cluster.Detail.Overview.02AE8EA0", defaultMessage: "修改删除保护已成功" }));
        }
      }
    }
  );

  return (
    <Card
      style={style}
      title={
      <h2 style={{ marginBottom: 0 }}>
          {intl.formatMessage({
          id: 'Dashboard.Detail.Overview.BasicInfo.ClusterInformation',
          defaultMessage: '集群信息'
        })}
        </h2>
      }>

      <Descriptions
        column={5}
        title={intl.formatMessage({
          id: 'Dashboard.Detail.Overview.BasicInfo.BasicClusterInformation',
          defaultMessage: '集群基本信息'
        })}>

        <Descriptions.Item
          label={intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.BasicInfo.ClusterName',
            defaultMessage: '集群名'
          })}>

          {clusterName}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.ResourceName',
            defaultMessage: '资源名'
          })}>

          {name}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.BasicInfo.Namespace',
            defaultMessage: '命名空间'
          })}>

          {namespace}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.ClusterMode',
            defaultMessage: '集群模式'
          })}>

          {MODE_MAP.get(mode)?.text || '-'}
        </Descriptions.Item>
        <Descriptions.Item
          label={intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.BasicInfo.ClusterStatus',
            defaultMessage: '集群状态'
          })}>

          <Tag color={statusItem.badgeStatus}>
            {statusItem === 'operating' ?
            <Text
              style={{ maxWidth: 110, color: '#d48806', fontSize: 12 }}
              ellipsis={{ tooltip: `${status}/${statusDetail}` }}>

                {statusItem.label}/{statusDetailItem.label}
              </Text> :

            statusItem.label
            }
          </Tag>
        </Descriptions.Item>

        <Descriptions.Item label={intl.formatMessage({ id: "src.pages.Cluster.Detail.Overview.8DB38279", defaultMessage: "删除保护" })}>
          <Checkbox
          // loading 态禁止操作，防止重复操作
          disabled={loading || status !== 'running'}
          defaultChecked={deletionProtection}
          onChange={(e) => {
            const body = {} as API.ParamPatchOBClusterParam;
            if (!e.target.checked) {
              body.removeDeletionProtection = e.target.checked;
            } else {
              body.addDeletionProtection = e.target.checked;
            }
            patchOBCluster(namespace, name, body);
          }} />

        </Descriptions.Item>
        <Descriptions.Item
          span={2}
          label={intl.formatMessage({
            id: 'OBDashboard.Detail.Overview.BasicInfo.Image',
            defaultMessage: '镜像'
          })}>

          {image}
        </Descriptions.Item>

        <Descriptions.Item
          span={2}
          label={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.RootPasswordSecret',
            defaultMessage: 'root 用户密码 Secret'
          })}>

          {rootPasswordSecret || '-'}
        </Descriptions.Item>
      </Descriptions>

      {monitor &&
      <Descriptions
        title={intl.formatMessage({
          id: 'Dashboard.Detail.Overview.BasicInfo.MonitoringConfiguration',
          defaultMessage: '监控配置'
        })}>

          <Descriptions.Item label="CPU">
            {monitor.resource.cpu}
          </Descriptions.Item>
          <Descriptions.Item label="Memory">
            {monitor.resource.memory}
          </Descriptions.Item>
          <Descriptions.Item
          label={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.Image',
            defaultMessage: '镜像'
          })}>

            {monitor.image}
          </Descriptions.Item>
        </Descriptions>
      }

      {backupVolume &&
      <Descriptions
        title={intl.formatMessage({
          id: 'Dashboard.Detail.Overview.BasicInfo.BackupVolumeConfiguration',
          defaultMessage: '备份卷配置'
        })}>

          <Descriptions.Item
          label={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.Address',
            defaultMessage: '地址'
          })}>

            {backupVolume.address}
          </Descriptions.Item>
          <Descriptions.Item
          label={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.Path',
            defaultMessage: '路径'
          })}>

            {backupVolume.path}
          </Descriptions.Item>
        </Descriptions>
      }
    </Card>);

}