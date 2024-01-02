export default [
  {
    path: '/',
    component: 'Layouts',
    routes: [
      {
        path: '/',
        component: 'Layouts/BasicLayout',
        name: '系统布局',
        routes: [
          {
            path: 'cluster',
            component: 'Cluster',
            name: '集群页',
          },
          {
            path:'cluster/new',
            component:'Cluster/New',
            name: '创建集群',
          },
          {
            path: 'overview',
            component: 'Overview',
            name: '系统概览页',
          },
          {
            path: '/',
            redirect: 'overview',
            name: '系统概览页',
          },
        ],
      },
      {
        path: 'cluster/:clusterId',
        component: 'Cluster/Detail',
        name: '集群详情',
        routes: [
          {
            path: 'overview',
            component: 'Cluster/Detail/Overview',
            name: '概览页',
          },
          {
            path: 'topo',
            component: 'Cluster/Detail/Topo',
            name: '集群拓扑图',
          },
          {
            path:'monitor',
            component:'Cluster/Detail/Monitor',
            name:'集群详情监控'
          },
          {
            path: '/cluster/:clusterId',
            redirect: 'overview',
            name: '概览页',
          },
        ],
      },
      {
        path: '/login',
        component: 'Login',
        name: '登录页',
      },
      {
        component: 'Error/404',
        name: '404 页面不存在',
      },
    ],
  },
];
