import { intl } from '@/utils/intl';
import { PageContainer } from '@ant-design/pro-components';
import { Row } from 'antd';

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
        <EventsTable />
        <NodesTable />
      </Row>
    </PageContainer>
  );
};

export default OverviewPage;
