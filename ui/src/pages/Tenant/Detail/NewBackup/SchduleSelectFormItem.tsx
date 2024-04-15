import { intl } from '@/utils/intl';
import { useUpdateEffect } from 'ahooks';
import type { FormInstance } from 'antd';
import { Form, Space } from 'antd';
import { clone } from 'lodash';
import type { ParamsType } from './ScheduleSelectComp';
import ScheduleSelectComp from './ScheduleSelectComp';

interface SchduleSelectFormItemProps {
  form: FormInstance<any>;
  scheduleValue: any;
  disable?: boolean;
}

export default function SchduleSelectFormItem({
  form,
  scheduleValue,
  disable = false,
}: SchduleSelectFormItemProps) {
  /**
   * When the scheduling period changes,
   * ensure that the backup data method
   * can be changed accordingly.
   */
  useUpdateEffect(() => {
    let newScheduleValue = clone(scheduleValue);
    scheduleValue.days.forEach((key: number) => {
      if (!scheduleValue[String(key)]) {
        newScheduleValue[key] = 'Full';
        form.setFieldValue('scheduleDates', newScheduleValue);
      }
    });
    Object.keys(scheduleValue).forEach((key) => {
      if (/^[\d]+$/.test(key) && !scheduleValue?.days.includes(Number(key))) {
        delete newScheduleValue[key];
        form.setFieldValue('scheduleDates', newScheduleValue);
      }
    });
  }, [scheduleValue]);

  return (
    <Space direction="vertical">
      <h3>
        {intl.formatMessage({
          id: 'Dashboard.Detail.NewBackup.SchduleSelectFormItem.SchedulingCycle',
          defaultMessage: '调度周期',
        })}
      </h3>
      <Form.Item
        rules={[
          () => ({
            validator: (_: any, value: ParamsType) => {
              if (!value.days.length) {
                return Promise.reject(
                  new Error(
                    intl.formatMessage({
                      id: 'Dashboard.Detail.NewBackup.SchduleSelectFormItem.SelectASchedulingCycle',
                      defaultMessage: '请选择调度周期',
                    }),
                  ),
                );
              }
              return Promise.resolve();
            },
          }),
        ]}
        name={['scheduleDates']}
      >
        <ScheduleSelectComp disable={disable} />
      </Form.Item>
    </Space>
  );
}
