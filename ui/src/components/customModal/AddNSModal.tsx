import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Form, Input, message } from 'antd';
import { useRef } from 'react';

import { createNameSpace } from '@/services';

import CustomModal from '.';

interface FormData {
  namespace: string;
}

export default function AddNSModal({
  visible,
  setVisible,
  successCallback,
}: API.CommonModalType) {
  const [form] = Form.useForm<FormData>();
  const newNamespace = useRef<string>('');
  const { run: createNS } = useRequest(createNameSpace, {
    manual: true,
    onSuccess: ({ successful }) => {
      if (successful) {
        message.success(
          intl.formatMessage({
            id: 'OBDashboard.components.customModal.AddNSModal.AddedSuccessfully',
            defaultMessage: '新增成功',
          }),
        );
        setVisible(false);
        if (successCallback) successCallback(newNamespace.current);
      }
    },
  });
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };
  const onCancel = () => setVisible(false);
  const onFinish = async (val: FormData) => {
    newNamespace.current = val.namespace;
    await createNS(val.namespace);
  };
  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'OBDashboard.components.customModal.AddNSModal.AddNamespace',
        defaultMessage: '新增命名空间',
      })}
      open={visible}
      onOk={handleSubmit}
      onCancel={onCancel}
    >
      <Form form={form} onFinish={onFinish}>
        <Form.Item
          label="namespace"
          name="namespace"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'OBDashboard.components.customModal.AddNSModal.EnterNamespace',
                defaultMessage: '请输入namespace!',
              }),
            },
          ]}
        >
          <Input
            placeholder={intl.formatMessage({
              id: 'OBDashboard.components.customModal.AddNSModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
