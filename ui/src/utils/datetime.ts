import { DATE_TIME_FORMAT } from '@/constants/datetime';
import {
  formatTime as formatTimeFromOBUtil,
  isNullValue,
} from '@oceanbase/util';

export function formatTime(
  value: number | string | undefined,
  format: string = DATE_TIME_FORMAT,
) {
  return isNullValue(value) ? '-' : formatTimeFromOBUtil(value, format);
}
