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
        status: 'Pending',
      },
    ],
    message: 'string',
    successful: true,
  },
  '/api/v1/obproxies/testns/testname': {
    data: {
      creationTime: 'string',
      image: 'oceanbase/oceanbase-cloud-native:4.2.0.0-101000032023091319',
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
      replicas: 10,
      resource: {
        cpu: 10,
        memory: 20,
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
      status: 'Pending',
    },
    message: 'string',
    successful: true,
  },
};
