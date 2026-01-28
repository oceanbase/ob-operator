import { intl } from '@/utils/intl';
import { range } from 'lodash';

export const SCHEDULE_TYPE_OPTIONS = [
  {
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Zhou',
      defaultMessage: '周',
    }),
    value: 'Weekly',
  },
  {
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Month',
      defaultMessage: '月',
    }),
    value: 'Monthly',
  },
];

export const INSPECTION_SCHEDULE_MODE_LIST = [
  {
    value: 'Monthly',
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Month',
      defaultMessage: '月',
    }),
  },

  {
    value: 'Weekly',
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Zhou',
      defaultMessage: '周',
    }),
  },

  {
    value: 'Dayly',
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Daily',
      defaultMessage: '日',
    }),
  },
];
export const WEEK_OPTIONS = [
  {
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.One',
      defaultMessage: '一',
    }),
    fullLabel: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Monday',
      defaultMessage: '星期一',
    }),
    value: 1,
  },

  {
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Ii',
      defaultMessage: '二',
    }),
    fullLabel: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Tuesday',
      defaultMessage: '星期二',
    }),

    value: 2,
  },

  {
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Three',
      defaultMessage: '三',
    }),
    fullLabel: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Wednesday',
      defaultMessage: '星期三',
    }),

    value: 3,
  },
  {
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Four',
      defaultMessage: '四',
    }),
    fullLabel: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Thursday',
      defaultMessage: '星期四',
    }),

    value: 4,
  },
  {
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Five',
      defaultMessage: '五',
    }),
    fullLabel: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Friday',
      defaultMessage: '星期五',
    }),

    value: 5,
  },
  {
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Six',
      defaultMessage: '六',
    }),
    fullLabel: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Saturday',
      defaultMessage: '星期六',
    }),

    value: 6,
  },
  {
    label: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Seven',
      defaultMessage: '日',
    }),
    fullLabel: intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Sunday',
      defaultMessage: '星期日',
    }),

    value: 7,
  },
];

export const WEEK_TEXT_MAP = new Map([
  [
    1,
    intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Monday',
      defaultMessage: '星期一',
    }),
  ],
  [
    2,
    intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Tuesday',
      defaultMessage: '星期二',
    }),
  ],
  [
    3,
    intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Wednesday',
      defaultMessage: '星期三',
    }),
  ],
  [
    4,
    intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Thursday',
      defaultMessage: '星期四',
    }),
  ],
  [
    5,
    intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Friday',
      defaultMessage: '星期五',
    }),
  ],
  [
    6,
    intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Saturday',
      defaultMessage: '星期六',
    }),
  ],
  [
    7,
    intl.formatMessage({
      id: 'Dashboard.src.constants.schedule.Sunday',
      defaultMessage: '星期日',
    }),
  ],
]);

export const MONTH_OPTIONS = range(1, 32).map((item) => ({
  label: item < 10 ? `0${item}` : item,
  fullLabel: item < 10 ? `0${item}` : item,
  value: item,
}));
