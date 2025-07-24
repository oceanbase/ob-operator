import { intl } from '@/utils/intl';

// 时间格式
export const TIME_FORMAT = 'HH:mm:ss';

// 不带秒信息的时间格式
export const TIME_FORMAT_WITHOUT_SECOND = 'HH:mm';

// 日期时间格式
export const DATE_TIME_FORMAT = 'YYYY-MM-DD HH:mm:ss';

export const DateSelectOption: Monitor.OptionType[] = [
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.CustomTime',
      defaultMessage: '自定义时间',
    }),
    value: 'custom',
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyMinutes',
      defaultMessage: '近30分钟',
    }),
    value: 1800000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyHour',
      defaultMessage: '近1小时',
    }),
    value: 3600000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyHours',
      defaultMessage: '近3小时',
    }),
    value: 10800000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyHours.1',
      defaultMessage: '近6小时',
    }),
    value: 21600000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyHours.2',
      defaultMessage: '近12小时',
    }),
    value: 43200000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.NearlyHours.3',
      defaultMessage: '近24小时',
    }),
    value: 86400000,
  },
  {
    label: intl.formatMessage({
      id: 'OBDashboard.Detail.Monitor.DataFilter.LastDays',
      defaultMessage: '近7天',
    }),
    value: 604800000,
  },
];
