import { intl } from '@/utils/intl';
import React from 'react';

const PlanDetail: React.FC = () => {
  return (
    <div>
      {intl.formatMessage({
        id: 'src.pages.Tenant.Detail.Sql.Plan.PlanDetailPageUnderConstruction',
        defaultMessage: '执行计划详情页面（建设中）',
      })}
    </div>
  );
};

export default PlanDetail;
