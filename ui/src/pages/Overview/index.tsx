import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { Col, Row } from 'antd';

import EventsTable from '../../components/EventsTable';
import NodesTable from './NodesTable';
import OverviewStatus from './OverviewStatus';

const OverviewPage: React.FC = () => {
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
        <OverviewStatus />
        <Col span={24}>
          <EventsTable />
        </Col>
        <NodesTable />
      </Row>
    </PageContainer>
  );
};

export default OverviewPage;
