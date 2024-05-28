import globalAxios, { AxiosInstance, AxiosPromise } from 'axios';
import {
  AlarmApiFactory,
  ClusterApiFactory,
  Configuration,
  InfoApiFactory,
  OBClusterApiFactory,
  OBTenantApiFactory,
  TerminalApiFactory,
  UserApiFactory,
  OBProxyApiFactory
} from './generated/index';

globalAxios.interceptors.response.use((res) => {
  return res.data;
});

const config = new Configuration({
  basePath: process.env.NODE_ENV === 'development' ? location.origin : '/',
  apiKey: () => document.cookie,
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
