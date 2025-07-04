import {
  MONTH_OPTIONS,
  SCHEDULE_TYPE_OPTIONS,
  WEEK_OPTIONS,
} from '@/constants/schedule';
import { Checkbox, Radio, theme } from '@oceanbase/design';
import styles from './index.less';

export type ScheduleMode = 'WEEK' | 'MONTH';

export interface Value {
  scheduleMode?: ScheduleMode;
  dayList?: number[];
}

export interface ScheduleModeAndDayListProps {
  value?: Value;
  onChange?: (value: Value) => void;
  // 调度周期
  SCHEDULE_MODE_LIST?: any[];
  // 默认调度周期
  initialMode?: string;
  type?: string;
}

// export interface ScheduleModeAndDayListInterface
//   extends React.FC<ScheduleModeAndDayListProps> {
//   validate: (rule: any, value: any, callback: any) => void;
// }

const ScheduleModeAndDayList: ScheduleModeAndDayListInterface = ({
  value = {},
  onChange = () => {},
  SCHEDULE_MODE_LIST,
  initialMode,
  type,
}) => {
  const { token } = theme.useToken();
  // 巡检调度规则，取消默认调度周期
  const { scheduleMode = initialMode || 'WEEK', dayList = [] } = value;

  return (
    <div className={styles.container}>
      <Radio.Group
        value={scheduleMode}
        onChange={(e) => {
          onChange({
            ...value,
            scheduleMode: e.target.value,
            dayList: [],
          });
        }}
      >
        {(SCHEDULE_MODE_LIST ? SCHEDULE_MODE_LIST : SCHEDULE_TYPE_OPTIONS).map(
          (item) => (
            <Radio.Button key={item.value} value={item.value}>
              {item.label}
            </Radio.Button>
          ),
        )}
      </Radio.Group>
      {/* 调度周期为日，样式不展示 */}
      {scheduleMode === 'MONTH' && (
        <ul
          className={styles.scheduleDayWrapper}
          style={{ backgroundColor: token.colorFillTertiary, marginBottom: 0 }}
        >
          {MONTH_OPTIONS.map((item) => (
            <li
              key={item.value}
              // 调度周期为月，选中日期数为 10 天时，disable 掉未选中的日期
              className={`${
                dayList.includes(item.value) ? styles.selected : ''
              } ${
                dayList.length === 10 && !dayList.includes(item.value)
                  ? styles.disabled
                  : ''
              }`}
              onClick={() => {
                onChange({
                  ...value,
                  dayList: dayList.includes(item.value)
                    ? dayList.filter((day) => day !== item.value)
                    : [...dayList, item.value],
                });
              }}
            >
              {item.fullLabel}
            </li>
          ))}
        </ul>
      )}
      {scheduleMode !== 'MONTH' && (
        <div style={{ paddingTop: 12 }}>
          {type !== 'inspection' && scheduleMode === 'WEEK' && (
            <Checkbox
              checked={dayList.length === 7}
              indeterminate={dayList.length > 0 && dayList.length < 7}
              onChange={(e) => {
                onChange({
                  ...value,
                  dayList: e.target.checked
                    ? WEEK_OPTIONS.map((item) => item.value)
                    : [],
                });
              }}
              style={{ marginRight: 8 }}
              className={styles.checkAll}
            >
              全选
            </Checkbox>
          )}
          {
            //  调度周期为月、周时，展示选项
            scheduleMode !== 'DAY'
              ? WEEK_OPTIONS.map((item) => (
                  <Checkbox
                    key={item.value}
                    checked={dayList.includes(item.value)}
                    onChange={() => {
                      onChange({
                        ...value,
                        dayList: dayList.includes(item.value)
                          ? dayList.filter((day) => day !== item.value)
                          : [...dayList, item.value],
                      });
                    }}
                    className={styles.check}
                  >
                    {item.fullLabel}
                  </Checkbox>
                ))
              : null
          }
        </div>
      )}
    </div>
  );
};

ScheduleModeAndDayList.validate = (rule, value, callback) => {
  if (value && !value.scheduleMode) {
    callback('请选择调度模式');
  }
  // 如果调度周期等于日时，跳出校验调度周期规则
  if (
    value &&
    value.scheduleMode !== 'DAY' &&
    (!value.dayList || value.dayList.length === 0)
  ) {
    callback('请选择调度周期');
  }

  callback();
};

export default ScheduleModeAndDayList;
