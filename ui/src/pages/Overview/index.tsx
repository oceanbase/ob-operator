import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { useAccess } from '@umijs/max';
import { Col, Row } from 'antd';

import EventsTable from '../../components/EventsTable';
import NodesTable from './NodesTable';
import OverviewStatus from './OverviewStatus';

const OverviewPage: React.FC = () => {
  const access = useAccess();
  return (
    <PageContainer
      header={{
        title: intl.formatMessage({
          id: 'dashboard.pages.Overview.Overview',
          defaultMessage: '概览',
        }),
      }}
    >
      <Row justify="start" gutter={[16, 16]}>
        {access.obclusterread || access.obclusterwrite ? (
          <OverviewStatus />
        ) : null}
        {access.systemread || access.systemwrite ? (
          <>
            <Col span={24}>
              <EventsTable overView={true} />
            </Col>
            <NodesTable type="overview" />
          </>
        ) : null}
      </Row>
    </PageContainer>
  );
};

export default OverviewPage;
