import type { RequestConfig } from '@umijs/max';
import { getLocale } from '@umijs/max';
import { message } from 'antd';
import dayjs from 'dayjs';
import localeData from 'dayjs/plugin/localeData';
import weekday from 'dayjs/plugin/weekday';
import { access } from './api';

dayjs.extend(weekday);
dayjs.extend(localeData);

interface ResponseStructure {
  successful: boolean;
}

export const request: RequestConfig = {
  timeout: 600000,

  errorConfig: {
    errorThrower: (res: ResponseStructure) => {
      const { successful } = res;
      if (!successful) {
        //throw error
        throw new Error(res);
      }
    },
    errorHandler: (err) => {
      console.log('errorHandler', err);

      const url = err?.config?.url || err?.request?.responseURL || '';

      if (err?.response?.status === 401) {
        location.href = '/#/login';
      } else if (url.includes('/api/web/oceanbase/report')) {
        console.log('Report data error, not showing message:', err);
      } else {
        message.error(err?.response?.data?.message || err.message);
      }
    },
  },
};
export const rootContainer = (element: JSX.Element) => {
  const locale = getLocale() || 'zh-CN';
  request.headers = {
    'Accept-Language': locale,
  };
  return <>{element}</>;
};

// If the user is not logged in, jump to the login page
// and refresh InitialState after successful login
export async function getInitialState() {
  const res = await Promise.all([
    access.getAccountInfo(),
    access.listAllPolicies(),
  ]);
  if (res[0].successful && res[1].successful) {
    return {
      accountInfo: res[0].data,
      policies: res[1].data,
    };
  }
  return {};
}
