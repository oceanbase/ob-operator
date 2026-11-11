import dayjs from 'dayjs';

export function formatEventTime(timestamp: number | null | undefined): string {
  if (typeof timestamp !== 'number' || !Number.isFinite(timestamp) || timestamp <= 0) return '-';
  const time = dayjs.unix(timestamp);
  return time.isValid() ? time.format('YYYY-MM-DD HH:mm:ss') : '-';
}
