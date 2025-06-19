import globalAxios, { AxiosInstance, AxiosPromise } from 'axios';
import { errorHandling } from './errorHandling';
import {
  AccessControlApiFactory,
  AlarmApiFactory,
  ClusterApiFactory,
  Configuration,
  InfoApiFactory,
  K8sClusterApiFactory,
  OBClusterApiFactory,
  OBProxyApiFactory,
  OBTenantApiFactory,
  TerminalApiFactory,
  UserApiFactory,
} from './generated/index';

globalAxios.interceptors.response.use((res) => {
  return res.data;
}, errorHandling);

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
export const K8sClusterApi = wrapper(K8sClusterApiFactory, config);
export const obcluster = wrapper(OBClusterApiFactory, config);
export const obtenant = wrapper(OBTenantApiFactory, config);
export const terminal = wrapper(TerminalApiFactory, config);
export const user = wrapper(UserApiFactory, config);
export const alert = wrapper(AlarmApiFactory, config);
export const obproxy = wrapper(OBProxyApiFactory, config);
export const access = wrapper(AccessControlApiFactory, config);
