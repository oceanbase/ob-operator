import { defineConfig } from '@umijs/max';
import routes from './routes';

export default defineConfig({
  antd: {},
  model: {},
  request: {},
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
  npmClient: 'yarn'
});
