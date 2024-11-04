export default [
  {
    path: '/',
    component: 'Layouts',
    routes: [
      {
        path: '/',
        component: 'Layouts/StatisticsLayout',
        name: '系统布局',
        routes: [
          {
            path: '/',
            component: 'Layouts/BasicLayout',
            name: '概览布局',
            routes: [
              {
                path: 'cluster',
                component: 'Cluster',
                name: '集群页',
              },

              {
                path: 'tenant',
                component: 'Tenant',
                name: '租户页',
              },

              {
                path: 'obproxy',
                component: 'OBProxy',
                name: 'obproxy',
              },

              {
                path: 'alert',
                component: 'Alert',
                name: '告警',
                routes: [
                  {
                    path: 'event',
                    component: 'Alert/Event',
                    name: '告警事件',
                  },
                  {
                    path: 'shield',
                    component: 'Alert/Shield',
                    name: '告警屏蔽',
                  },
                  {
                    path: 'rules',
                    component: 'Alert/Rules',
                    name: '告警规则',
                  },
                  {
                    path: 'channel',
                    component: 'Alert/Channel',
                    name: '告警通道',
                  },
                  {
                    path: 'subscriptions',
                    component: 'Alert/Subscriptions',
                    name: '告警推送',
                  },
                  {
                    path: '/alert',
                    redirect: 'event',
                  },
                ],
              },
              {
                path: 'access',
                component: 'Access',
                name: '权限控制',
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
            path: 'tenant/new',
            component: 'Tenant/New',
            name: '创建租户',
          },
          {
            path: 'cluster/new',
            component: 'Cluster/New',
            name: '创建集群',
          },
          {
            path: 'obproxy/new',
            component: 'OBProxy/New',
            name: '创建obproxy',
          },
          {
            path: 'cluster/:ns/:name/:clusterName',
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
                path: 'monitor',
                component: 'Cluster/Detail/Monitor',
                name: '集群详情监控',
              },
              {
                path: 'tenant',
                component: 'Cluster/Detail/Tenant',
                name: '集群下的租户',
              },
              {
                path: 'connection',
                component: 'Cluster/Detail/Connection',
                name: '连接集群',
              },
              {
                path: '/cluster/:ns/:name/:clusterName',
                redirect: 'overview',
                name: '概览页',
              },
            ],
          },
          {
            path: 'tenant/:ns/:name/:tenantName',
            component: 'Tenant/Detail',
            name: '租户详情',
            routes: [
              {
                path: 'overview',
                component: 'Tenant/Detail/Overview',
                name: '概览页',
              },
              {
                path: 'topo',
                component: 'Tenant/Detail/Topo',
                name: '租户拓扑图',
              },
              {
                path: 'backup',
                component: 'Tenant/Detail/Backup',
                name: '租户备份',
              },
              {
                path: 'backup/new',
                component: 'Tenant/Detail/NewBackup',
                name: '新建租户备份',
              },
              {
                path: 'monitor',
                component: 'Tenant/Detail/Monitor',
                name: '租户详情监控',
              },
              {
                path: 'connection',
                component: 'Tenant/Detail/Connection',
                name: '连接租户',
              },
              {
                path: '/tenant/:ns/:name/:tenantName',
                redirect: 'overview',
                name: '概览页',
              },
            ],
          },
          {
            path: 'obproxy/:ns/:name',
            component: 'OBProxy/Detail',
            name: 'obproxy详情',
            routes: [
              {
                path: 'overview',
                component: 'OBProxy/Detail/Overview',
                name: '概览页',
              },
              {
                path: 'monitor',
                component: 'OBProxy/Detail/Monitor',
                name: 'obproxy详情监控',
              },
              {
                path: '/obproxy/:ns/:name',
                redirect: 'overview',
                name: '概览页',
              },
            ],
          },
        ],
      },
      {
        path: '/login',
        component: 'Login',
        name: '登录页',
      },
      {
        path: '/reset',
        component: 'ResetPwd',
        name: '重置密码',
      },
      {
        component: 'Error/404',
        name: '404 页面不存在',
      },
    ],
  },
];
