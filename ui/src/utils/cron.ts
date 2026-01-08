/**
 * Cron expression parsing utility
 * Supports standard cron format (5 fields) and extended format (6 fields)
 */

import { intl } from '@/utils/intl';

export interface CronObject {
  minute: string;
  hour: string;
  dayOfMonth: string;
  month: string;
  dayOfWeek: string;
  year?: string;
  nextRun?: string;
  prevRun?: string;
  description?: string;
}

export interface CronParseResult {
  success: boolean;
  data?: CronObject;
  error?: string;
}

// 常量定义
const CRON_FIELD_PATTERNS = {
  minute: /^[\d\*,\-\/]+$/,
  hour: /^[\d\*,\-\/]+$/,
  dayOfMonth: /^[\d\*,\-\/]+$/,
  month: /^[\d\*,\-\/]+$/,
  dayOfWeek: /^[\d\*,\-\/]+$/,
  year: /^[\d\*,\-\/]+$/,
} as const;

const getWeekDay = (index: number): string => {
  const keys = [
    'src.utils.cron.Sunday',
    'src.utils.cron.Monday',
    'src.utils.cron.Tuesday',
    'src.utils.cron.Wednesday',
    'src.utils.cron.Thursday',
    'src.utils.cron.Friday',
    'src.utils.cron.Saturday',
  ];

  const defaultMessages = [
    intl.formatMessage({ id: 'src.utils.EE6CFC53', defaultMessage: '周日' }),
    intl.formatMessage({ id: 'src.utils.AEF107F1', defaultMessage: '周一' }),
    intl.formatMessage({ id: 'src.utils.F86B45D3', defaultMessage: '周二' }),
    intl.formatMessage({ id: 'src.utils.6BE72650', defaultMessage: '周三' }),
    intl.formatMessage({ id: 'src.utils.B4A71F61', defaultMessage: '周四' }),
    intl.formatMessage({ id: 'src.utils.CF6B8D0F', defaultMessage: '周五' }),
    intl.formatMessage({ id: 'src.utils.D96B2392', defaultMessage: '周六' }),
  ];
  return intl.formatMessage({
    id: keys[index % 7],
    defaultMessage: defaultMessages[index % 7],
  });
};

// 字段验证器
const FIELD_VALIDATORS = [
  (val: string) => CRON_FIELD_PATTERNS.minute.test(val),
  (val: string) => CRON_FIELD_PATTERNS.hour.test(val),
  (val: string) => CRON_FIELD_PATTERNS.dayOfMonth.test(val),
  (val: string) => CRON_FIELD_PATTERNS.month.test(val),
  (val: string) => CRON_FIELD_PATTERNS.dayOfWeek.test(val),
] as const;

/**
 * 验证 cron 表达式格式
 */
function isValidCronFormat(cronExpression: string): boolean {
  const parts = cronExpression.trim().split(/\s+/);

  // 检查字段数量
  if (parts.length < 5 || parts.length > 6) {
    return false;
  }

  // 验证每个字段
  const validators = [...FIELD_VALIDATORS];
  if (parts.length === 6) {
    validators.push((val: string) => CRON_FIELD_PATTERNS.year.test(val));
  }

  return parts.every((part, index) => validators[index](part));
}

/**
 * 解析 cron 表达式字段
 */
function parseCronFields(cronExpression: string): CronObject {
  const parts = cronExpression.trim().split(/\s+/);

  const cronObject: CronObject = {
    minute: parts[0],
    hour: parts[1],
    dayOfMonth: parts[2],
    month: parts[3],
    dayOfWeek: parts[4],
  };

  if (parts.length === 6) {
    cronObject.year = parts[5];
  }

  return cronObject;
}

/**
 * 生成字段描述
 */
function generateFieldDescription(
  value: string,
  fieldType: keyof typeof CRON_FIELD_PATTERNS,
): string {
  if (value === '*') {
    const descriptions = {
      minute: intl.formatMessage({
        id: 'src.utils.cron.EveryMinute',
        defaultMessage: '每分钟',
      }),
      hour: intl.formatMessage({
        id: 'src.utils.cron.EveryHour',
        defaultMessage: '每小时',
      }),
      dayOfMonth: intl.formatMessage({
        id: 'src.utils.cron.EveryDay',
        defaultMessage: '每天',
      }),
      month: intl.formatMessage({
        id: 'src.utils.cron.EveryMonth',
        defaultMessage: '每月',
      }),
      dayOfWeek: intl.formatMessage({
        id: 'src.utils.cron.EveryWeek',
        defaultMessage: '每周',
      }),
      year: intl.formatMessage({
        id: 'src.utils.cron.EveryYear',
        defaultMessage: '每年',
      }),
    };
    return descriptions[fieldType];
  }

  if (value.includes('/')) {
    const interval = value.split('/')[1];
    const units = {
      minute: intl.formatMessage({
        id: 'src.utils.cron.Minute',
        defaultMessage: '分钟',
      }),
      hour: intl.formatMessage({
        id: 'src.utils.cron.Hour',
        defaultMessage: '小时',
      }),
      dayOfMonth: intl.formatMessage({
        id: 'src.utils.cron.Day',
        defaultMessage: '天',
      }),
      month: intl.formatMessage({
        id: 'src.utils.cron.Month',
        defaultMessage: '个月',
      }),
      dayOfWeek: intl.formatMessage({
        id: 'src.utils.cron.Week',
        defaultMessage: '周',
      }),
      year: intl.formatMessage({
        id: 'src.utils.cron.Year',
        defaultMessage: '年',
      }),
    };
    return intl.formatMessage(
      {
        id: 'src.utils.cron.EveryInterval',
        defaultMessage: '每{interval}{unit}',
      },
      { interval, unit: units[fieldType] },
    );
  }

  if (value.includes(',')) {
    const values = value.split(',');
    if (fieldType === 'dayOfWeek') {
      const days = values.map((d) => getWeekDay(parseInt(d)));
      return intl.formatMessage(
        { id: 'src.utils.cron.OnDays', defaultMessage: '在{days}' },
        {
          days: days.join(
            intl.formatMessage({
              id: 'src.utils.cron.Comma',
              defaultMessage: '、',
            }),
          ),
        },
      );
    }
    return intl.formatMessage(
      { id: 'src.utils.cron.OnValue', defaultMessage: '在{value}' },
      { value },
    );
  }

  if (value.includes('-')) {
    return intl.formatMessage(
      { id: 'src.utils.cron.OnValue', defaultMessage: '在{value}' },
      { value },
    );
  }

  if (fieldType === 'dayOfWeek') {
    return intl.formatMessage(
      { id: 'src.utils.cron.OnDay', defaultMessage: '在{day}' },
      { day: getWeekDay(parseInt(value)) },
    );
  }

  return intl.formatMessage(
    { id: 'src.utils.cron.OnValue', defaultMessage: '在{value}' },
    { value },
  );
}

/**
 * 生成 cron 表达式的可读描述
 */
function generateDescription(cronExpression: string): string {
  const parts = cronExpression.trim().split(/\s+/);
  const fieldTypes: (keyof typeof CRON_FIELD_PATTERNS)[] = [
    'minute',
    'hour',
    'dayOfMonth',
    'month',
    'dayOfWeek',
  ];

  const descriptions = fieldTypes.map((fieldType, index) =>
    generateFieldDescription(parts[index], fieldType),
  );

  return descriptions.join('');
}

/**
 * 计算下次执行时间（简化版本）
 */
function calculateNextRunTime(): string {
  // 这里可以实现更复杂的计算逻辑
  // 目前返回当前时间作为示例
  return new Date().toISOString();
}

/**
 * 计算上次执行时间（简化版本）
 */
function calculatePrevRunTime(): string {
  // 这里可以实现更复杂的计算逻辑
  // 目前返回一小时前的时间作为示例
  const oneHourAgo = new Date();
  oneHourAgo.setHours(oneHourAgo.getHours() - 1);
  return oneHourAgo.toISOString();
}

/**
 * 将 cron 表达式转换为对象
 */
export function parseCronExpression(cronExpression: string): CronParseResult {
  try {
    if (!isValidCronFormat(cronExpression)) {
      return {
        success: false,
        error: intl.formatMessage({
          id: 'src.utils.cron.InvalidCronFormat',
          defaultMessage: '无效的 cron 表达式格式',
        }),
      };
    }

    const cronObject = parseCronFields(cronExpression);
    cronObject.description = generateDescription(cronExpression);
    cronObject.nextRun = calculateNextRunTime();
    cronObject.prevRun = calculatePrevRunTime();

    return {
      success: true,
      data: cronObject,
    };
  } catch (error) {
    return {
      success: false,
      error: error instanceof Error ? error.message : 'Unknown error',
    };
  }
}

/**
 * 验证 cron 表达式是否有效
 */
export function isValidCronExpression(cronExpression: string): boolean {
  return isValidCronFormat(cronExpression);
}

/**
 * 获取 cron 表达式的下一次执行时间
 */
export function getNextRunTime(cronExpression: string): string | null {
  if (!isValidCronExpression(cronExpression)) {
    return null;
  }
  return calculateNextRunTime();
}

/**
 * 获取 cron 表达式的上一次执行时间
 */
export function getPrevRunTime(cronExpression: string): string | null {
  if (!isValidCronExpression(cronExpression)) {
    return null;
  }
  return calculatePrevRunTime();
}

/**
 * 将对象转换回 cron 表达式
 */
export function objectToCronExpression(cronObject: CronObject): string {
  const parts = [
    cronObject.minute,
    cronObject.hour,
    cronObject.dayOfMonth,
    cronObject.month,
    cronObject.dayOfWeek,
  ];

  if (cronObject.year) {
    parts.push(cronObject.year);
  }

  return parts.join(' ');
}

/**
 * 获取 cron 表达式的详细信息
 */
export function getCronDetails(cronExpression: string) {
  const parseResult = parseCronExpression(cronExpression);

  if (!parseResult.success || !parseResult.data) {
    return {
      isValid: false,
      error: parseResult.error,
    };
  }

  const { data } = parseResult;

  return {
    isValid: true,
    expression: cronExpression,
    fields: {
      minute: data.minute,
      hour: data.hour,
      dayOfMonth: data.dayOfMonth,
      month: data.month,
      dayOfWeek: data.dayOfWeek,
      year: data.year,
    },
    nextRun: data.nextRun,
    prevRun: data.prevRun,
    description: data.description,
  };
}

/**
 * 预设的常用 cron 表达式
 */
export const CRON_PRESETS = {
  daily: '0 0 * * *',
  dailyNoon: '0 12 * * *',
  weeklyMonday: '0 0 * * 1',
  weeklyMondayNoon: '0 12 * * 1',
  monthly: '0 0 1 * *',
  every15Minutes: '*/15 * * * *',
  every30Minutes: '*/30 * * * *',
  hourly: '0 * * * *',
  every2Hours: '0 */2 * * *',
  every6Hours: '0 */6 * * *',
  weekdays: '0 0 * * 1,3,5',
  weekdaysNoon: '0 12 * * 1,3,5',
  biweekly: '0 0 1,15 * *',
  yearly: '0 0 1 1 *',
  specificYear: '0 0 1 1 * 2024',
} as const;

/**
 * 验证并格式化 cron 表达式
 */
export function validateAndFormatCron(cronExpression: string) {
  const isValid = isValidCronExpression(cronExpression);

  if (!isValid) {
    return {
      isValid: false,
      error: intl.formatMessage({
        id: 'src.utils.cron.InvalidCronExpression',
        defaultMessage: '无效的 cron 表达式',
      }),
      suggestions: getSuggestions(cronExpression),
    };
  }

  const details = getCronDetails(cronExpression);

  return {
    isValid: true,
    expression: cronExpression,
    details,
    formatted: {
      description: details.description,
      nextRun: details.nextRun
        ? new Date(details.nextRun).toLocaleString()
        : intl.formatMessage({
            id: 'src.utils.cron.CannotCalculate',
            defaultMessage: '无法计算',
          }),
      prevRun: details.prevRun
        ? new Date(details.prevRun).toLocaleString()
        : intl.formatMessage({
            id: 'src.utils.cron.CannotCalculate',
            defaultMessage: '无法计算',
          }),
    },
  };
}

/**
 * 获取 cron 表达式的建议修复
 */
function getSuggestions(cronExpression: string): string[] {
  const suggestions: string[] = [];
  const parts = cronExpression.trim().split(/\s+/);

  if (parts.length < 5) {
    suggestions.push(
      intl.formatMessage({
        id: 'src.utils.cron.StandardCronRequires5Fields',
        defaultMessage:
          '标准 cron 表达式需要 5 个字段：分钟 小时 日期 月份 星期',
      }),
    );
    suggestions.push(
      intl.formatMessage({
        id: 'src.utils.cron.Example',
        defaultMessage: '示例：0 0 * * *',
      }),
    );
  } else if (parts.length > 6) {
    suggestions.push(
      intl.formatMessage({
        id: 'src.utils.cron.CronSupportsMax6Fields',
        defaultMessage: 'cron 表达式最多支持 6 个字段（包含年份）',
      }),
    );
  }

  // 检查常见错误
  const commonErrors = [
    {
      pattern: /60/,
      message: intl.formatMessage({
        id: 'src.utils.cron.MinuteFieldCannotExceed59',
        defaultMessage: '分钟字段不能超过 59',
      }),
    },
    {
      pattern: /25/,
      message: intl.formatMessage({
        id: 'src.utils.cron.HourFieldCannotExceed23',
        defaultMessage: '小时字段不能超过 23',
      }),
    },
    {
      pattern: /32/,
      message: intl.formatMessage({
        id: 'src.utils.cron.DayFieldCannotExceed31',
        defaultMessage: '日期字段不能超过 31',
      }),
    },
    {
      pattern: /13/,
      message: intl.formatMessage({
        id: 'src.utils.cron.MonthFieldCannotExceed12',
        defaultMessage: '月份字段不能超过 12',
      }),
    },
  ];

  commonErrors.forEach(({ pattern, message }) => {
    if (pattern.test(cronExpression)) {
      suggestions.push(message);
    }
  });

  return suggestions;
}

/**
 * 表单验证辅助函数
 */
export function validateCronInForm(cronExpression: string) {
  const result = validateAndFormatCron(cronExpression);

  if (!result.isValid) {
    return {
      validateStatus: 'error' as const,
      help: result.error,
      suggestions: result.suggestions,
    };
  }

  return {
    validateStatus: 'success' as const,
    help:
      result.formatted?.description ||
      intl.formatMessage({
        id: 'src.utils.cron.ValidCronExpression',
        defaultMessage: '有效的 cron 表达式',
      }),
  };
}
