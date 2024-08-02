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
  timeout: 10000,

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
      if (err?.response?.status === 401) {
        location.href = '/#/login';
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
