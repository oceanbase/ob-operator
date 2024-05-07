import {
  ClusterApiFactory,
  Configuration,
  InfoApiFactory,
  OBClusterApiFactory,
  OBTenantApiFactory,
  TerminalApiFactory,
  UserApiFactory,
  AlarmApiFactory
} from './generated/index';

const config = new Configuration({
  basePath:
    process.env.NODE_ENV === 'development' ? 'http://localhost:8000' : '/',
  apiKey: () => document.cookie,
  baseOptions: {
    withCredentials: true,
  }
});
export const info = InfoApiFactory(config);
export const cluster = ClusterApiFactory(config);
export const obcluster = OBClusterApiFactory(config);
export const obtenant = OBTenantApiFactory(config);
export const terminal = TerminalApiFactory(config);
export const user = UserApiFactory(config);
export const alert = AlarmApiFactory(config);
