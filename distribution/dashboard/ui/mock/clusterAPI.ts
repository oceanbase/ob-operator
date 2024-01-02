export default {
  '/cluster/events': {
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
};
