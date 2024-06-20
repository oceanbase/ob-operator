import { intl } from '@/utils/intl';
import { message } from 'antd';
import globalAxios, { AxiosInstance, AxiosPromise } from 'axios';
import {
  AlarmApiFactory,
  ClusterApiFactory,
  Configuration,
  InfoApiFactory,
  OBClusterApiFactory,
  OBProxyApiFactory,
  OBTenantApiFactory,
  TerminalApiFactory,
  UserApiFactory,
} from './generated/index';

globalAxios.interceptors.response.use(
  (res) => {
    return res.data;
  },
  (error) => {
    if (error?.response?.status === 401) {
      message.warning(
        intl.formatMessage({
          id: 'src.api.2CA64FC6',
          defaultMessage: '登陆已过期',
        }),
      );
      location.href = '/#/login';
    } else {
      message.error(error?.response?.data?.message || error.message);
    }
    return Promise.reject(error.response);
  },
);

const config = new Configuration({
  basePath: location.origin,
  baseOptions: {
    withCredentials: true,
  },
});

type factoryFunction<T> = (
  configuration?: Configuration | undefined,
  basePath?: string | undefined,
  axios?: AxiosInstance | undefined,
) => T;

const wrapper = <T>(
  f: factoryFunction<T>,
  ...args: Parameters<factoryFunction<T>>
): PromiseWrapperType<T> => {
  return f(...args) as any;
};
type PromiseWrapperType<T> = {
  [K in keyof T]: T[K] extends (...args: infer P) => AxiosPromise<infer R>
    ? (...args: P) => Promise<R>
    : never;
};

export const info = wrapper(InfoApiFactory, config);
export const cluster = wrapper(ClusterApiFactory, config);
export const obcluster = wrapper(OBClusterApiFactory, config);
export const obtenant = wrapper(OBTenantApiFactory, config);
export const terminal = wrapper(TerminalApiFactory, config);
export const user = wrapper(UserApiFactory, config);
export const alert = wrapper(AlarmApiFactory, config);
export const obproxy = wrapper(OBProxyApiFactory, config);
