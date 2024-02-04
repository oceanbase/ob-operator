export default {
  '/obclusters/statistic': {
    data: [
      {
        status: 'deleting',
        count: 0,
      },
      {
        status: 'operating',
        count: 1,
      },
      {
        status: 'running',
        count: 3,
      },
    ],
    message: '',
    successful: true,
  },

  'POST /obclusters': {
    data: {
      namespace: 'oceanbase',
      name: 'test',
    },
    message: '',
    successful: true,
  },

  '/obclusters': {
    data: [
      {
        namespace: 'oceanbase',
        name: 'test',
        status: 'running',
        createTime: '2023-08-20 11:11:11',
        image: 'oceanbasedev/oceanbase-cn:4.1.0.1-test',
        metrics: {
          cpuPercent: 30,
          diskPercent: 40,
          memoryPercent: 20,
        },
        topology: [
          {
            zone: 'zone1',
            replicas: 2,
            name: 'string',
            namespace: 'string',
            rootService: 'string',
            status: 'string',
            nodeSelector: [
              {
                key: 'ob.zone',
                value: 'zone1',
              },
            ],
            observers: [
              {
                address: 'string',
                metrics: {
                  cpuPercent: 0,
                  diskPercent: 0,
                  memoryPercent: 0,
                },
                name: 'string',
                namespace: 'string',
                status: 'string',
              },
            ],
          },
        ],
      },
    ],
    message: '',
    successful: true,
  },

  '/api/v1/obclusters/namespace/test/name/test': {
    data: {
      namespace: 'oceanbase',
      createTime: 'string',
      name: 'test',
      status: 'operating',
      image: 'oceanbasedev/oceanbase-cn:4.1.0.1-test',
      metrics: {
        cpuPercent: 10,
        diskPercent: 20,
        memoryPercent: 30,
      },
      topology: [
        {
          name: 'obcluster-test-obzone1',
          namespace: 'topology',
          zone: 'zone1',
          replicas: '2',
          status: 'deleting',
          rootService: '1.1.1.1',
          nodeSelector: [
            {
              key: 'string',
              value: 'string',
            },
          ],
          observers: [
            {
              name: 'obcluster-obzone1-xxxxxx',
              namespace: 'observers',
              status: 'running',
              address: '1.1.1.1',
              metrics: {
                cpuPercent: '40',
                memoryPercent: '50',
                diskPercent: '60',
              },
            },
          ],
        },
        {
          name: 'obcluster-test-obzone2',
          namespace: 'topology2',
          zone: 'zone2',
          replicas: '1',
          status: 'deleting',
          rootService: '1.1.1.1',
          nodeSelector: [
            {
              key: 'string',
              value: 'string',
            },
          ],
          observers: [
            {
              name: 'obcluster-obzone2-xxxxxx',
              namespace: 'observers2',
              status: 'deleting',
              address: '1.1.12.2',
              metrics: {
                cpuPercent: '40',
                memoryPercent: '50',
                diskPercent: '60',
              },
            },
          ],
        },
      ],
    },
    message: '',
    successful: true,
  },

  'POST /obclusters/namespace/oceanbase/name/test': {
    data: '',
    message: '升级成功',
    successful: true,
  },

  'POST /obclusters/namespace/oceanbase/name/test/obzones/zone1/scale': {
    data: '',
    message: '',
    successful: true,
  },

  'POST /obclusters/namespace/oceanbase/name/test/obzones': {
    data: '',
    message: '',
    successful: true,
  },

  'DELETE /obclusters/namespace/oceanbase/name/test': {
    data: '',
    message: '删除成功',
    successful: true,
  },

  'DELETE /obclusters/namespace/oceanbase/name/test/obzones/zone1': {
    data: '',
    message: '',
    successful: true,
  },

  'DELETE /observers/namespace/oceanbase/name/test': {
    data: '',
    message: '',
    successful: true,
  },
};
