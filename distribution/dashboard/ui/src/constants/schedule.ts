import { range } from 'lodash';

export const SCHEDULE_TYPE_OPTIONS = [
  {
    label: '周',
    value: 'Weekly',
  },
  {
    label: '月',
    value: 'Monthly',
  },
];

export const WEEK_OPTIONS = [
  {
    label: '一',
    fullLabel: '星期一',
    value: 1,
  },

  {
    label: '二',
    fullLabel: '星期二',

    value: 2,
  },

  {
    label: '三',
    fullLabel: '星期三',

    value: 3,
  },
  {
    label: '四',
    fullLabel: '星期四',

    value: 4,
  },
  {
    label: '五',
    fullLabel: '星期五',

    value: 5,
  },
  {
    label: '六',
    fullLabel: '星期六',

    value: 6,
  },
  {
    label: '日',
    fullLabel: '星期日',

    value: 7,
  },
];

export const WEEK_TEXT_MAP = new Map([
  [1, '星期一'],
  [2, '星期二'],
  [3, '星期三'],
  [4, '星期四'],
  [5, '星期五'],
  [6, '星期六'],
  [7, '星期日'],
]);

export const MONTH_OPTIONS = range(1, 32).map((item) => ({
  label: item < 10 ? `0${item}` : item,
  fullLabel: item < 10 ? `0${item}` : item,
  value: item,
}));
