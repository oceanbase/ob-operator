export default {
  '/api/v1/obproxies': {
    data: [
      {
        creationTime: 'string',
        image: 'string',
        name: 'testname',
        namespace: 'testns',
        obCluster: {
          name: 'string',
          namespace: 'string',
        },
        proxyClusterName: 'string',
        replicas: 0,
        serviceIp: 'string',
        status: 'string',
      },
    ],
    message: 'string',
    successful: true,
  },
  '/api/v1/obproxies/testns/testname': {
    data: {
      creationTime: 'string',
      image: 'string',
      name: 'testname',
      namespace: 'testns',
      obCluster: {
        name: 'string',
        namespace: 'string',
      },
      parameters: [
        {
          key: 'string',
          value: 'string',
        },
      ],
      pods: [
        {
          containers: [
            {
              image: 'string',
              limits: {
                cpu: 0,
                memory: 0,
              },
              name: 'string',
              ports: [0],
              ready: true,
              requests: {
                cpu: 0,
                memory: 0,
              },
              restartCount: 0,
              startTime: 'string',
            },
          ],
          message: 'string',
          name: 'string',
          namespace: 'string',
          nodeName: 'string',
          podIP: 'string',
          reason: 'string',
          startTime: 'string',
          status: 'string',
        },
      ],
      proxyClusterName: 'string',
      proxySysSecret: 'string',
      replicas: 0,
      resource: {
        cpu: 0,
        memory: 0,
      },
      service: {
        clusterIP: 'string',
        externalIP: 'string',
        name: 'string',
        namespace: 'string',
        ports: [
          {
            name: 'string',
            port: 0,
            targetPort: 0,
          },
        ],
        type: 'string',
      },
      serviceIp: 'string',
      status: 'string',
    },
    message: 'string',
    successful: true,
  },
};
