import type { RequestConfig } from '@umijs/max';
import { message } from 'antd';
import { getLocale } from '@umijs/max';

// 错误处理方案： 错误类型
enum ErrorShowType {}

interface ResponseStructure {
  successful: boolean;
  showType?: ErrorShowType;
}

export const request: RequestConfig = {
  timeout: 10000,

  errorConfig: {
    // 基于responseInterceptors实现，触发条件：data.success === false
    errorThrower: (res: ResponseStructure) => {
      const { successful } = res;
      if (!successful) {
        //throw error
        throw new Error(res);
      }
    },
    // 会 catch errorThrower 抛出的错误
    errorHandler: (err) => {
      console.log('errorHandler', err);
      if (err?.response?.status === 401) {
        location.href = '/#/login';
      } else {
        message.error(err?.response?.data?.message || err.message);
      }
    },
  },

  // 请求拦截器
  requestInterceptors: [
    // (url, options) => {
    //   if (localStorage.getItem('token')) {
    //     options.headers.Authorization = localStorage.getItem('token');
    //   }
    //   return { url, options };
    // },
  ],

  // 响应拦截器
  responseInterceptors: [
    // (response) => {
    //   console.log('response', response);
    //   if (response.status === 401) {
    //     message.info(response.data.message);
    //     location.href = '/login';
    //   }
    //   return response;
    // },
  ],
};
export const rootContainer = (element: JSX.Element) => {
  const locale = getLocale() || 'zh-CN';
  request.headers = {
    'Accept-Language': locale,
  };
  return <>{element}</>;
};