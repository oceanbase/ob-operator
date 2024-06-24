export default {
  '/api/v1/cluster/events': {
    data: [
      {
        namespace: 'oceanbase',
        type: 'NORMAL', //NORMAL, WARNING or nil
        count: 0,
        firstOccur: 0,
        lastSeen: 0,
        reason: 'Unhealthy',
        object: 'pod/obcluster-obzone1-xerijei',
        message: 'pod recreate',
      },
    ],
    message: '',
    successful: true,
  },

  '/cluster/nodes': {
    data: [
      {
        info: {
          name: 'sqaappnoxdnv62s2011161204050.sa128',
          status: 'Ready',
          roles: 'control-plane,master',
          uptime: '28d',
          version: 'v1.23.6+k3s1',
          internalIp: '11.161.204.50',
          externalIp: '',
          os: 'Alibaba Group Enterprise Linux Server 7.2 (Paladin)',
          kernel: '4.9.151-015.ali3000.alios7.x86_64',
          cri: 'docker://18.6.1',
        },
        resource: {
          cpuTotal: 32, //cpu总量
          cpuUsed: 20.5, //已使用
          memoryTotal: 12341234123, //单位字节
          memoryUsed: 41234123,
          // no disk displayed in resource, since it's possible to use remote disk
        },
      },
    ],
    message: '',
    successful: true,
  },

  '/api/v1/obclusters': {
    data: [
      {
        clusterId: 0,
        clusterName: 'string',
        createTime: 0,
        image: 'string',
        mode: 'NORMAL',
        name: 'string',
        namespace: 'string',
        status: 'string',
        statusDetail: 'string',
        topology: [
          {
            affinities: [
              {
                key: 'string',
                type: 'NODE',
                value: 'string',
              },
            ],
            name: 'string',
            namespace: 'string',
            nodeSelector: [
              {
                key: 'string',
                value: 'string',
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
                statusDetail: 'string',
              },
            ],
            replicas: 0,
            rootService: 'string',
            status: 'string',
            statusDetail: 'string',
            tolerations: [
              {
                key: 'string',
                value: 'string',
              },
            ],
            zone: 'string',
          },
        ],
        uid: 'string',
      },
      {
        clusterId: 0,
        clusterName: 'string1',
        createTime: 0,
        image: 'string1',
        mode: 'NORMAL',
        name: 'string1',
        namespace: 'string1',
        status: 'string1',
        statusDetail: 'string1',
        topology: [
          {
            affinities: [
              {
                key: 'string1',
                type: 'NODE',
                value: 'string1',
              },
            ],
            name: 'string1',
            namespace: 'string1',
            nodeSelector: [
              {
                key: 'string1',
                value: 'string1',
              },
            ],
            observers: [
              {
                address: 'string1',
                metrics: {
                  cpuPercent: 0,
                  diskPercent: 0,
                  memoryPercent: 0,
                },
                name: 'string1',
                namespace: 'string1',
                status: 'string1',
                statusDetail: 'string1',
              },
            ],
            replicas: 0,
            rootService: 'string1',
            status: 'string1',
            statusDetail: 'string1',
            tolerations: [
              {
                key: 'string1',
                value: 'string1',
              },
            ],
            zone: 'string1',
          },
        ],
        uid: 'string1',
      },
    ],
    message: 'string1',
    successful: true,
  },
};
