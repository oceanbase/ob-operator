import { obtenant } from '@/api';
import { STATUS_LIST } from '@/constants';
import { intl } from '@/utils/intl';
import { findByValue } from '@oceanbase/util';
import { useRequest } from 'ahooks';
import { Card, Checkbox, Descriptions, Tag, message } from 'antd';
import { isEmpty } from 'lodash';

export default function BasicInfo({
  info = {},
  source = {},
  loading,
  style,
  name,
  ns,
}: API.TenantBasicInfo & { style?: React.CSSProperties; loading: boolean }) {
  const InfoConfig = {
    name: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.ResourceName',
      defaultMessage: '资源名',
    }),
    namespace: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.Namespace',
      defaultMessage: '命名空间',
    }),
    tenantName: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.TenantName',
      defaultMessage: '租户名',
    }),
    clusterResourceName: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.ClusterName',
      defaultMessage: '集群名',
    }),
    tenantRole: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.TenantRole',
      defaultMessage: '租户角色',
    }),
    charset: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.CharacterSet',
      defaultMessage: '字符集',
    }),
    status: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.Status',
      defaultMessage: '状态',
    }),
    unitNumber: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.NumberOfUnits',
      defaultMessage: 'unit 数量',
    }),
    locality: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.ReplicaDistribution',
      defaultMessage: '副本分布',
    }),
    primaryZone: 'PrimaryZone',
  };
  const SourceConfig = {
    primaryTenant: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.MasterTenant',
      defaultMessage: '主租户',
    }),
    restoreType: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.RecoverySource',
      defaultMessage: '恢复源',
    }),
    archiveSource: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.ArchiveSource',
      defaultMessage: '存档来源',
    }),
    bakDataSource: intl.formatMessage({
      id: 'Dashboard.Detail.Overview.BasicInfo.DataSource',
      defaultMessage: '数据源',
    }),
    until: 'until',
  };

  const checkSource = (source: API.Source) => {
    Object.keys(source).forEach((key: keyof API.Source) => {
      if (source[key]) return true;
    });
    return false;
  };

  const { runAsync: patchTenant, loading: patchTenantLoading } = useRequest(
    obtenant.patchTenant,
    {
      manual: true,
      onSuccess: (res) => {
        if (res.successful) {
          message.success(
            intl.formatMessage({
              id: 'src.pages.Tenant.Detail.Overview.892E62C7',
              defaultMessage: '修改删除保护已成功',
            }),
          );
        }
      },
    },
  );
  const { deletionProtection } = info;

  return (
    <Card
      loading={loading}
      title={
        <h2 style={{ marginBottom: 0 }}>
          {intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.TenantBasicInformation',
            defaultMessage: '租户信息',
          })}
        </h2>
      }
      style={style}
    >
      <Descriptions column={5}>
        {Object.keys(InfoConfig).map((key, index) => {
          const statusItem = findByValue(STATUS_LIST, info.status);
          return (
            <Descriptions.Item key={index} label={InfoConfig[key]}>
              {key !== 'status' ? (
                info[key]
              ) : !isEmpty(statusItem) ? (
                <Tag color={statusItem.badgeStatus}>{statusItem.label}</Tag>
              ) : (
                '-'
              )}
            </Descriptions.Item>
          );
        })}

        <Descriptions.Item
          label={intl.formatMessage({
            id: 'src.pages.Tenant.Detail.Overview.EE772326',
            defaultMessage: '删除保护',
          })}
        >
          <Checkbox
            // loading 态禁止操作，防止重复操作
            disabled={patchTenantLoading}
            defaultChecked={deletionProtection}
            onChange={(e) => {
              const body = {} as API.ParamPatchTenant;
              if (!e.target.checked) {
                body.removeDeletionProtection = true;
              } else {
                body.addDeletionProtection = true;
              }
              patchTenant(ns, name, body);
            }}
          />
        </Descriptions.Item>
      </Descriptions>

      {checkSource(source) && (
        <Descriptions
          title={intl.formatMessage({
            id: 'Dashboard.Detail.Overview.BasicInfo.TenantResources',
            defaultMessage: '租户资源',
          })}
        >
          {Object.keys(SourceConfig).map((key, index) => (
            <Descriptions.Item label={SourceConfig[key]} key={index}>
              {source[key]}
            </Descriptions.Item>
          ))}
        </Descriptions>
      )}
    </Card>
  );
}
