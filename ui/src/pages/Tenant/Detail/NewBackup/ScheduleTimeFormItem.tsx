import { intl } from '@/utils/intl';
import { Form, Space, TimePicker } from 'antd';

export default function ScheduleTimeFormItem({
  disable = false,
}: {
  disable?: boolean;
}) {
  return (
      <Space direction="vertical">
        <h3>
          {intl.formatMessage({
            id: 'Dashboard.Detail.NewBackup.ScheduleTimeFormItem.SchedulingTime',
            defaultMessage: '调度时间',
          })}
        </h3>
        <Form.Item
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.Detail.NewBackup.ScheduleTimeFormItem.SelectASchedulingTime',
                defaultMessage: '请选择调度时间',
              }),
            },
          ]}
          name={'scheduleTime'}
        >
          <TimePicker format="HH:mm" disabled={disable} />
        </Form.Item>
      </Space>
  );
}
