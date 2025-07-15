import { defineConfig } from '@umijs/max';
import routes from './routes';

export default defineConfig({
  antd: {},
  access: {},
  model: {},
  request: {},
  initialState: {},
  favicons: ['/logo.png'],
  title: 'OceanBase Dashboard',
  layout: false,
  mock: false,
  locale: {
    default: 'zh-CN',
    baseSeparator: '-',
  },
  routes,
  history: { type: 'hash' },
  npmClient: 'yarn',
  jsMinifier: 'terser',
  proxy: {
    '/api/v1': {
      target: 'http://11.161.204.49:30555',
    },
  },
});
