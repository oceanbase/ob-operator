export default {
  'POST /api/v1/alarm/alert/alerts': {
    data: [
      {
        description: 'string',
        endsAt: 0,
        fingerprint: 'string',
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
        serverity: 'critical',
        startsAt: 0,
        status: {
          inhibitedBy: ['string'],
          silencedBy: ['string'],
          state: 'active',
        },
        summary: 'string',
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
        endsAt: 0,
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
        startsAt: 0,
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
};
