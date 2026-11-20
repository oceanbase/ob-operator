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
  // Keep clients from combining scripts from different Dashboard releases.
  hash: true,
  locale: {
    default: 'zh-CN',
    baseSeparator: '-',
  },
  routes,
  history: { type: 'hash' },
  npmClient: 'yarn',
  jsMinifier: 'terser',
});
