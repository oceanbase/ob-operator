import { intl } from '@/utils/intl';
import { Form } from 'antd';

export const CustomFormItem = (prop: any) => {
  const { label, message } = prop;
  return (
    <Form.Item
      {...prop}
      rules={[
        {
          required: true,
          message:
            message ||
            intl.formatMessage(
              {
                id: 'Dashboard.Cluster.New.Observer.EnterLabel',
                defaultMessage: '请输入{{label}}',
              },
              { label: label },
            ),
        },
      ]}
    >
      {prop.children}
    </Form.Item>
  );
};
