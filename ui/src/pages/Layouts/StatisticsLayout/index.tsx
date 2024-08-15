import { Outlet } from '@umijs/max';
import { useEffect } from 'react';

import { CHECK_STORAGE_INTERVAL } from '@/constants';
import { isReportTimeExpired, reportPollData } from '@/utils/helper';
import { useModel, useNavigate } from '@umijs/max';

export default function StatisticsLayout() {
  const { reportDataInterval } = useModel('global');
  const navigate = useNavigate();
  const { initialState } = useModel('@@initialState');
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

  useEffect(() => {
    if (initialState?.accountInfo?.needReset) {
      navigate('/reset');
    }
  }, [initialState]);

  return <Outlet />;
}
