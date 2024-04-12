import { WEEK_TEXT_MAP } from '@/constants/schedule';
import { intl } from '@/utils/intl';
import { Form, Radio, Space } from 'antd';
import type { FormInstance } from 'antd/lib/form';

interface BacMethodsListProps {
  scheduleValue?: {
    mode: 'Monthly' | 'Weekly';
    days: number[];
  };
  disable?: boolean;
  form?: FormInstance<any>;
}

export default function BakMethodsList({
  scheduleValue,
  disable = false,
  form,
}: BacMethodsListProps) {
  const dataSource = scheduleValue || form?.getFieldValue('scheduleDates');

  return (
    <Space direction="vertical" style={{ marginBottom: 24, width: 440 }}>
      <h3>
        {intl.formatMessage({
          id: 'Dashboard.Detail.NewBackup.BakMethodsList.BackupData',
          defaultMessage: '备份数据方式',
        })}
      </h3>
      <p>
        {intl.formatMessage({
          id: 'Dashboard.Detail.NewBackup.BakMethodsList.ConfigureAtLeastOneFull',
          defaultMessage: '至少配置 1 个全量备份',
        })}
      </p>
      {dataSource?.days
        .sort((pre: number, cur: number) => pre - cur)
        .map((day: number, index: number) => (
          <Form.Item
            name={['scheduleDates', day]}
            label={
              dataSource?.mode === 'Monthly' ? day : WEEK_TEXT_MAP.get(day)
            }
            labelCol={{span:6}}
            wrapperCol={{span:18}}
            style={{ marginBottom: 0 }}
            key={index}
          >
            <Radio.Group disabled={disable} defaultValue="Full">
              <Radio value="Full">
                {intl.formatMessage({
                  id: 'Dashboard.Detail.NewBackup.BakMethodsList.FullQuantity',
                  defaultMessage: '全量',
                })}
              </Radio>
              <Radio value="Incremental">
                {intl.formatMessage({
                  id: 'Dashboard.Detail.NewBackup.BakMethodsList.Increment',
                  defaultMessage: '增量',
                })}
              </Radio>
            </Radio.Group>
          </Form.Item>
        ))}
    </Space>
  );
}
