export default {
  'POST /api/v1/alarm/alert/alerts': {
    data: [
      {
        description: 'string',
        endsAt: 0,
        fingerprint: 'string1',
        instance: {
          obcluster: 'testobcluster',
          type: 'obcluster',
        },
        labels: [
          {
            key: 'string',
            value: 'string',
          },
          {
            key: 'string1',
            value: 'string1',
          },
          {
            key: 'string2',
            value: 'string2',
          },
        ],
        rule: 'string',
        severity: 'critical',
        startsAt: 1716543006,
        status: {
          inhibitedBy: ['string'],
          silencedBy: ['string'],
          state: 'suppressed',
        },
        summary: 'string1',
        updatedAt: 0,
      },
      {
        description: 'string',
        endsAt: 0,
        fingerprint: 'string2',
        instance: {
          obcluster: 'testobcluster',
          type: 'obcluster',
        },
        labels: [
          {
            key: 'string',
            value: 'string',
          },
          {
            key: 'string1',
            value: 'string1',
          },
          {
            key: 'string2',
            value: 'string2',
          },
        ],
        rule: 'string',
        severity: 'info',
        startsAt: 1716548006,
        status: {
          inhibitedBy: ['string'],
          silencedBy: ['string'],
          state: 'unprocessed',
        },
        summary: 'string2',
        updatedAt: 0,
      },
      {
        description: 'string',
        endsAt: 0,
        fingerprint: 'string3',
        instance: {
          obcluster: 'testobcluster',
          type: 'obcluster',
        },
        labels: [
          {
            key: 'string',
            value: 'string',
          },
          {
            key: 'string1',
            value: 'string1',
          },
          {
            key: 'string2',
            value: 'string2',
          },
        ],
        rule: 'string',
        severity: 'warning',
        startsAt: 1716548606,
        status: {
          inhibitedBy: ['string'],
          silencedBy: ['string'],
          state: 'active',
        },
        summary: 'string3',
        updatedAt: 0,
      },
    ],
    message: 'string',
    successful: true,
  },
  'POST /api/v1/alarm/silence/silencers': {
    data: [
      {
        comment: 'string',
        createdBy: 'string',
        endsAt: 1716543006,
        id: 'string',
        instance: {
          obcluster: 'string',
          observer: 'string',
          obtenant: 'string',
          obzone: 'string',
          type: 'obcluster',
        },
        matchers: [
          {
            isEqual: true,
            isRegex: true,
            name: 'string',
            value: 'string',
          },
        ],
        startsAt: 1716543006,
        status: {
          state: 'expired',
        },
        updatedAt: 0,
      },
      {
        comment: 'string',
        createdBy: 'string',
        endsAt: 1716543006,
        id: 'string',
        instance: {
          obcluster: 'string',
          observer: 'string',
          obtenant: 'string',
          obzone: 'string',
          type: 'obcluster',
        },
        matchers: [
          {
            isEqual: true,
            isRegex: true,
            name: 'string',
            value: 'string',
          },
        ],
        startsAt: 1716543006,
        status: {
          state: 'pending',
        },
        updatedAt: 0,
      },
      {
        comment: 'string1',
        createdBy: 'string1',
        endsAt: 1716548006,
        id: 'string1',
        instance: {
          obcluster: 'string',
          observer: 'string',
          obtenant: 'string',
          obzone: 'string',
          type: 'obcluster',
        },
        matchers: [
          {
            isEqual: true,
            isRegex: true,
            name: 'string',
            value: 'string',
          },
        ],
        startsAt: 1716548006,
        status: {
          state: 'active',
        },
        updatedAt: 0,
      },
    ],
    message: 'string',
    successful: true,
  },
  'POST /api/v1/alarm/route/routes': {
    data: [
      {
        aggregateLabels: ['string'],
        groupInterval: 0,
        groupWait: 0,
        id: 'string',
        matchers: [
          {
            isEqual: true,
            isRegex: true,
            name: 'string',
            value: 'string',
          },
        ],
        receiver: 'string',
        repeatInterval: 0,
      },
    ],
    message: 'string',
    successful: true,
  },
  'POST /api/v1/alarm/receiver/receivers': {
    data: [
      {
        config: 'string',
        name: 'string',
        type: 'dingtalk',
      },
    ],
    message: 'string',
    successful: true,
  },
  'POST /api/v1/alarm/rule/rules': {
    data: [
      {
        description: 'string',
        duration: 0,
        evaluationTime: 0,
        health: 'unknown',
        instanceType: 'obcluster',
        keepFiringFor: 0,
        labels: {
          key: 'string',
          value: 'string',
        },
        lastError: 'string',
        lastEvaluation: 0,
        name: 'string',
        query: 'string',
        severity: 'critical',
        state: 'active',
        summary: 'string',
        type: 'builtin',
      },
      {
        description: 'string',
        duration: 0,
        evaluationTime: 0,
        health: 'unknown',
        instanceType: 'obcluster',
        keepFiringFor: 0,
        labels: {
          key: 'string',
          value: 'string',
        },
        lastError: 'string',
        lastEvaluation: 0,
        name: 'string',
        query: 'string',
        severity: 'caution',
        state: 'active',
        summary: 'string',
        type: 'customized',
      },
      {
        description: 'string',
        duration: 0,
        evaluationTime: 0,
        health: 'unknown',
        instanceType: 'obcluster',
        keepFiringFor: 0,
        labels: {
          key: 'string',
          value: 'string',
        },
        lastError: 'string',
        lastEvaluation: 0,
        name: 'string',
        query: 'string',
        severity: 'warning',
        state: 'active',
        summary: 'string',
        type: 'builtin',
      },
    ],
    message: 'string',
    successful: true,
  },
  'GET /api/v1/alarm/silence/silencers/string': {
    data: {
      comment: 'string',
      createdBy: 'string',
      endsAt: 1716282320833,
      id: 'string',
      instances: [
        {
          obcluster: 'string',
          observer: 'string',
          obtenant: 'string',
          obzone: 'string',
          type: 'obcluster',
        },
      ],
      rules: ['string'],
      matchers: [
        {
          isRegex: true,
          name: 'string',
          value: 'string',
        },
      ],
      startsAt: 0,
      status: {
        state: 'active',
      },
      updatedAt: 0,
    },
    message: 'string',
    successful: true,
  },
  'GET /api/v1/alarm/rule/rules/string': {
    data: {
      description: 'string',
      duration: 0,
      evaluationTime: 0,
      health: 'unknown',
      instanceType: 'obcluster',
      keepFiringFor: 0,
      labels: [
        {
          key: 'string',
          value: 'string',
        },
      ],
      lastError: 'string',
      lastEvaluation: 0,
      name: 'string',
      query: 'string',
      severity: 'critical',
      state: 'active',
      summary: 'string',
      type: 'builtin',
    },
    message: 'string',
    successful: true,
  },
  'GET /api/v1/alarm/receiver/receivers/string': {
    data: {
      config: 'string',
      name: 'string',
      type: 'discord',
    },
    message: 'string',
    successful: true,
  },
  'GET /api/v1/alarm/route/routes/string': {
    data: {
      aggregateLabels: ['string'],
      groupInterval: 0,
      groupWait: 0,
      id: 'string',
      matchers: [
        {
          isRegex: true,
          name: 'string',
          value: 'string',
        },
      ],
      receiver: 'string',
      repeatInterval: 0,
    },
    message: 'string',
    successful: true,
  },
};
