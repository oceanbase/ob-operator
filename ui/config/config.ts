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
  mock: false,
  layout: false,
  locale: {
    default: 'zh-CN',
    baseSeparator: '-',
  },
  routes,
  history: { type: 'hash' },
  npmClient: 'yarn',
  proxy: {
    '/api': {
      target: 'http://11.161.204.4:18083',
      changeOrigin: true,
      ws: true,
    }
  }
});
