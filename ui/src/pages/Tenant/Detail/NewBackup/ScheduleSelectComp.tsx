import {
  INSPECTION_SCHEDULE_MODE_LIST,
  MONTH_OPTIONS,
  SCHEDULE_TYPE_OPTIONS,
  WEEK_OPTIONS,
} from '@/constants/schedule';
import { intl } from '@/utils/intl';
import { Checkbox, Radio } from 'antd';
import styles from './index.less';

export type ParamsType = {
  mode: 'Weekly' | 'Monthly';
  days: number[];
};

interface ScheduleSelectCompProps {
  value: ParamsType;
  onChange: (val: ParamsType) => void;
  disable?: boolean;
}

export default function ScheduleSelectComp({
  value: params,
  onChange,
  disable,
  type,
}: ScheduleSelectCompProps) {
  const handleSelectDay = (day: number) => {
    onChange({
      ...params,
      days: params.days.includes(day)
        ? params.days.filter((curDay) => curDay !== day)
        : [...params.days, day],
    });
  };

  return (
    <div>
      <Radio.Group
        options={
          type === 'inspection'
            ? INSPECTION_SCHEDULE_MODE_LIST
            : SCHEDULE_TYPE_OPTIONS
        }
        onChange={(e) => {
          onChange({
            mode: e.target.value,
            days: [],
          });
        }}
        value={params.mode}
        optionType="button"
        buttonStyle="solid"
        disabled={disable}
      />
      {params.mode !== 'Dayly' && (
        <ul className={styles.container}>
          {(params.mode === 'Weekly' ? WEEK_OPTIONS : MONTH_OPTIONS).map(
            (item, index) => (
              <li
                className={`${
                  params.days.includes(item.value) ? styles.selected : ''
                } ${
                  params.days.length === 10 && !params.days.includes(item.value)
                    ? styles.disabled
                    : ''
                }`}
                onClick={() => handleSelectDay(item.value)}
                key={index}
                style={disable ? { pointerEvents: 'none' } : {}}
              >
                {item.label}
              </li>
            ),
          )}
          {params.mode === 'Weekly' && (
            <Checkbox
              disabled={disable}
              checked={params.days.length === 7}
              onChange={(e) =>
                onChange({
                  ...params,
                  days: e.target.checked
                    ? WEEK_OPTIONS.map((item) => item.value)
                    : [],
                })
              }
              indeterminate={params.days.length > 0 && params.days.length < 7}
              style={{ marginLeft: 10, marginTop: 8 }}
            >
              {intl.formatMessage({
                id: 'Dashboard.Detail.NewBackup.ScheduleSelectComp.SelectAll',
                defaultMessage: '全选',
              })}
            </Checkbox>
          )}
        </ul>
      )}
    </div>
  );
}
