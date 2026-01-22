import ClusterSelect from '@/components/ClusterSelect';
import TenantSelect from '@/components/TenantSelect';
import { STATUS_LIST } from '@/constants';
import DetailLayout from '@/pages/Layouts/DetailLayout';
import { getClusterDetailReq } from '@/services';
import { getTenant } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { ApartmentOutlined } from '@ant-design/icons';
import type { MenuItem } from '@oceanbase/design/es/BasicLayout';
import { IconFont } from '@oceanbase/ui';
import { findByValue } from '@oceanbase/util';
import { useAccess, useParams } from '@umijs/max';
import { useRequest } from 'ahooks';
import { Badge } from 'antd';

export default () => {
  const params = useParams();
  const { ns, name, tenantName } = params;
  const access = useAccess();
  const menus: MenuItem[] = [
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.Overview',
        defaultMessage: '概览',
      }),
      link: `/tenant/${ns}/${name}/${tenantName}`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.TopologyDiagram',
        defaultMessage: '拓扑图',
      }),
      link: `/tenant/${ns}/${name}/${tenantName}/topo`,
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Tenant.Detail.AACEF9DE',
        defaultMessage: '备份/恢复',
      }),
      link: `/tenant/${ns}/${name}/${tenantName}/backup`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.PerformanceMonitoring',
        defaultMessage: '性能监控',
      }),
      key: 'monitor',
      link: `/tenant/${ns}/${name}/${tenantName}/monitor`,
    },
    {
      title: intl.formatMessage({
        id: 'Dashboard.Tenant.Detail.Connection1',
        defaultMessage: '连接租户',
      }),
      key: 'connection',
      link: `/tenant/${ns}/${name}/${tenantName}/connection`,
      accessible: access.obclusterwrite,
    },
    {
      title: intl.formatMessage({
        id: 'src.pages.Tenant.Detail.Sql.SqlAnalysis',
        defaultMessage: 'SQL 分析',
      }),
      link: `/tenant/${ns}/${name}/${tenantName}/sql`,
    },
  ];

  const { data: tenantDetailResponse } = useRequest(getTenant, {
    defaultParams: [{ ns: ns!, name: name! }],
    onSuccess: ({ data, successful }) => {
      if (successful) {
        getClusterDetail({ ns: ns!, name: data?.info?.clusterResourceName });
      }
    },
  });

  const tenantDetail = tenantDetailResponse?.data;

  const { data: clusterDetail, run: getClusterDetail } = useRequest(
    getClusterDetailReq,
    {
      manual: true,
    },
  );

  const statusDetailItem = findByValue(STATUS_LIST, tenantDetail?.info?.status);

  return (
    <DetailLayout
      menus={menus}
      subSideSelectKey="tenant"
      sideHeader={
        clusterDetail?.info?.clusterId ? (
          <div style={{ position: 'relative', paddingLeft: '16px' }}>
            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                position: 'relative',
              }}
            >
              <div
                style={{
                  position: 'relative',
                  display: 'flex',
                  alignItems: 'center',
                }}
              >
                <ApartmentOutlined style={{ fontSize: '16px' }} />
                <div
                  style={{
                    position: 'absolute',
                    left: '8px',
                    top: '16px',
                    width: '1px',
                    height: '21px',
                    borderLeft: '1px dashed #d9d9d9',
                  }}
                />
              </div>
              <ClusterSelect
                valueProp="id"
                value={`${clusterDetail?.info?.clusterName}:${clusterDetail?.info?.clusterId}`}
                bordered={false}
                showInnerStandby={false}
                suffixIcon={null}
                optionLabelRender={(item) => item?.name}
              />
            </div>

            <div
              style={{
                display: 'flex',
                alignItems: 'center',
                marginLeft: '24px',
                position: 'relative',
              }}
            >
              <div
                style={{
                  position: 'absolute',
                  left: '-17px',
                  top: '16px',
                  width: '20px',
                  height: '1px',
                  borderTop: '1px dashed #d9d9d9',
                }}
              />
              <div
                style={{
                  position: 'relative',
                  display: 'flex',
                  alignItems: 'center',
                }}
              >
                <IconFont type="tenant" style={{ fontSize: '16px' }} />
              </div>
              <TenantSelect
                valueProp="id"
                value={tenantDetail?.info?.tenantName}
                bordered={false}
                showStandby={true}
                showInnerStandby={false}
                suffixIcon={null}
                optionLabelRender={(item) => item?.name}
                clusterResourceName={tenantDetail?.info?.clusterResourceName}
              />
            </div>

            <div style={{ marginLeft: '27px', marginTop: '2px' }}>
              <Badge
                color={statusDetailItem.badgeStatus}
                text={statusDetailItem.label}
              />
            </div>
          </div>
        ) : null
      }
    />
  );
};
