import { intl } from '@/utils/intl';
import type { FormItemProps } from 'antd';
import { Form } from 'antd';

export const CustomFormItem = (prop: FormItemProps & { message?: string }) => {
  const { label, message } = prop;
  return (
    <Form.Item
      rules={[
        {
          required: true,
          message:
            message ||
            (typeof label === 'string'
              ? intl.formatMessage(
                  {
                    id: 'Dashboard.Cluster.New.Observer.EnterLabel',
                    defaultMessage: '请输入{{label}}',
                  },
                  { label: label as string },
                )
              : intl.formatMessage({
                  id: 'src.components.CustomFormItem.2C6315A1',
                  defaultMessage: '请输入',
                })),
        },
      ]}
      {...prop}
    >
      {prop.children}
    </Form.Item>
  );
};
