import { intl } from '@/utils/intl';
import { InfoCircleFilled } from '@ant-design/icons';
import { token } from '@oceanbase/design';
import { useUpdateEffect } from 'ahooks';
import type { FormInstance } from 'antd';
import { Form, Space } from 'antd';
import { clone } from 'lodash';
import type { ParamsType } from './ScheduleSelectComp';
import ScheduleSelectComp from './ScheduleSelectComp';

interface SchduleSelectFormItemProps {
  form: FormInstance<API.NewBackupForm>;
  scheduleValue: OBTenant.ScheduleDates;
  disable?: boolean;
  type?: string;
}

export default function SchduleSelectFormItem({
  form,
  scheduleValue,
  disable = false,
  type,
}: SchduleSelectFormItemProps) {
  /**
   * When the scheduling period changes,
   * ensure that the backup data method
   * can be changed accordingly.
   */
  useUpdateEffect(() => {
    const newScheduleValue = clone(scheduleValue);
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
      {type === 'inspection' ? (
        <h4>
          {intl.formatMessage({
            id: 'Dashboard.Detail.NewBackup.SchduleSelectFormItem.SchedulingCycle',
            defaultMessage: '调度周期',
          })}
        </h4>
      ) : (
        <h3>
          {intl.formatMessage({
            id: 'Dashboard.Detail.NewBackup.SchduleSelectFormItem.SchedulingCycle',
            defaultMessage: '调度周期',
          })}
        </h3>
      )}

      <Form.Item
        rules={[
          () => ({
            validator: (_: unknown, value: ParamsType) => {
              if (!value.days.length && scheduleValue?.mode !== 'Dayly') {
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
        extra={
          type === 'inspection' &&
          scheduleValue?.mode === 'Monthly' &&
          // 调度周期为月时，且选中日期为31号时，出现提示
          (scheduleValue?.days?.find((item) => item === 31) ? (
            <Space>
              <InfoCircleFilled style={{ color: token.colorPrimary }} />
              <span>
                {intl.formatMessage({
                  id: 'src.pages.Tenant.Detail.NewBackup.2B9E9733',
                  defaultMessage:
                    '若调度的当月没有 31 号，则对该月的最后 1 天进行调度',
                })}
              </span>
            </Space>
          ) : (
            intl.formatMessage({
              id: 'src.pages.Tenant.Detail.NewBackup.31C3220E',
              defaultMessage: '调度周期为月时，最多只能选择 10 天',
            })
          ))
        }
      >
        <ScheduleSelectComp disable={disable} type="inspection" />
      </Form.Item>
    </Space>
  );
}
