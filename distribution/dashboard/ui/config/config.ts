import { defineConfig } from '@umijs/max';
import routes from './routes';

export default defineConfig({
  antd: {},
  access: {},
  model: {},
  initialState: {},
  request: {},
  favicons: ['/logo.png'],
  title: 'OceanBase Dashboard',
  layout: false,
  locale: {
    default: 'zh-CN',
    baseSeparator: '-',
  },
  routes,
  history: { type: 'hash' },
  npmClient: 'yarn',
  mock: false,
  proxy: {
    '/api/v1': {
      target: 'http://11.161.204.4:18081',
    },
  },
});
