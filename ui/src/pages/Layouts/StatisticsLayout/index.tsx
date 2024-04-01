import { Outlet } from '@umijs/max';
import { useEffect } from 'react';

import { CHECK_STORAGE_INTERVAL } from '@/constants';
import { isReportTimeExpired, reportPollData } from '@/utils/helper';
import { useModel } from '@umijs/max';

export default function StatisticsLayout() {
  const { reportDataInterval } = useModel('global');
  useEffect(() => {
    const lastReportTimestamp = Number(localStorage.getItem('lastReportTime'));
    if (!lastReportTimestamp || isReportTimeExpired(lastReportTimestamp)) {
      reportPollData();
    }

    if (!reportDataInterval.current) {
      reportDataInterval.current = setInterval(() => {
        const reportTimestamp = Number(localStorage.getItem('lastReportTime'));
        if (isReportTimeExpired(reportTimestamp)) {
          reportPollData();
        }
      }, CHECK_STORAGE_INTERVAL);
    }

    return () => {
      clearInterval(reportDataInterval.current);
    };
  }, []);
  return <Outlet />;
}
