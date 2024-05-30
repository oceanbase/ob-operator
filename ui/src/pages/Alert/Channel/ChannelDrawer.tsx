import { alert } from '@/api';
import AlertDrawer from '@/components/AlertDrawer';
import { Alert } from '@/type/alert';
import { QuestionCircleOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import type { DrawerProps } from 'antd';
import { Button, Form, Input, Select, Space } from 'antd';
import { useEffect } from 'react';

/**
 * ChannelDrawer has three states: creation,editing and pure display
 * Editing state and pure display state can be switched between each other
 */
interface ChannelDrawerProps extends DrawerProps {
  status: Alert.DrawerStatus;
  name?: string;
  setStatus?: (status: Alert.DrawerStatus) => void;
}

const { TextArea } = Input;

export default function ChannelDrawer({
  status,
  name,
  onClose,
  setStatus,
  ...props
}: ChannelDrawerProps) {
  const [form] = Form.useForm();
  const { run: getReceiver } = useRequest(alert.getReceiver, {
    manual: true,
    onSuccess: ({ successful, data }) => {
      if (successful) {
        form.setFieldsValue({
          ...data,
        });
      }
    },
  });
  const { data: listReceiverTemplatesRes } = useRequest(
    alert.listReceiverTemplates,
  );
  const { data: listReceiversRes, run: getListReceivers } = useRequest(
    alert.listReceivers,
    {
      manual: true,
    },
  );
  const receiverNames = listReceiversRes?.data.map((receiver) => receiver.name);
  const listReceiverTemplates = listReceiverTemplatesRes?.data;
  const Footer = () => {
    return (
      <div>
        <Space>
          <Button
            onClick={
              status === 'display' && setStatus
                ? () => setStatus('edit')
                : () => form.submit()
            }
            type="primary"
          >
            {status === 'display' ? '编辑' : '提交'}
          </Button>
          <Button onClick={onClose}>取消</Button>
        </Space>
      </div>
    );
  };

  useEffect(() => {
    if (status !== 'create' && name) {
      getReceiver(name);
    }
    if (status === 'create') {
      getListReceivers();
    }
  }, [status, name]);

  return (
    <AlertDrawer
      title="告警通道配置"
      footer={<Footer />}
      destroyOnClose={true}
      onSubmit={() => form.submit()}
      onClose={onClose}
      {...props}
    >
      <Form form={form} preserve={false} layout="vertical">
        <Form.Item
          wrapperCol={{ span: 12 }}
          label="通道名称"
          rules={[
            {
              required: true,
              message: '请输入',
            },
            {
              validator: (_, value) => {
                if (
                  status === 'create' &&
                  receiverNames?.some((receiver) => receiver === value)
                ) {
                  return Promise.reject('告警通道已存在，请重新输入');
                }
                return Promise.resolve();
              },
            },
          ]}
          name={'name'}
        >
          {status === 'display' ? (
            <p>{form.getFieldValue('name') || '-'}</p>
          ) : (
            <Input disabled={status !== 'create'} placeholder="请输入" />
          )}
        </Form.Item>
        <Form.Item
          wrapperCol={{ span: 12 }}
          label="通道类型"
          rules={[
            {
              required: true,
              message: '请选择',
            },
          ]}
          name={'type'}
        >
          {status === 'display' ? (
            <p>{form.getFieldValue('type') || '-'} </p>
          ) : (
            <Select
              placeholder="请选择"
              options={listReceiverTemplates?.map((template) => ({
                value: template.type,
                label: template.type,
              }))}
            />
          )}
        </Form.Item>
        <Form.Item noStyle dependencies={['type']}>
          {({ setFieldValue, getFieldValue }) => {
            const type = getFieldValue('type');
            const template = listReceiverTemplates?.find(
              (item) => item.type === type,
            );
            if (template) setFieldValue('config', template.template);
            return (
              <Form.Item
                name={'config'}
                rules={[
                  {
                    required: true,
                    message: '请输入',
                  },
                ]}
                label={
                  <div>
                    <span>通道配置 </span>
                    <QuestionCircleOutlined />
                  </div>
                }
              >
                {status === 'display' ? (
                  <pre>{form.getFieldValue('config') || '-'}</pre>
                ) : (
                  <TextArea rows={18} placeholder="请输入" />
                )}
              </Form.Item>
            );
          }}
        </Form.Item>
      </Form>
    </AlertDrawer>
  );
}
