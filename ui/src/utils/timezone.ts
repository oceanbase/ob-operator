// 获取浏览器当前时区信息
export const getTimezoneInfo = () => {
  const date = new Date();
  const timezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
  const offset = -date.getTimezoneOffset() / 60;

  return {
    name: timezone,
    offset,
    offsetMinutes: date.getTimezoneOffset(),
    formatted: `${timezone} (UTC${offset >= 0 ? '+' : ''}${offset})`,
  };
};

// 获取时区名称
export const getTimezoneName = () => {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
};

// 获取时区偏移量（小时）
export const getTimezoneOffset = () => {
  return -new Date().getTimezoneOffset() / 60;
};
