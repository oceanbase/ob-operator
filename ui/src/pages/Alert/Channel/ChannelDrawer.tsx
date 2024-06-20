import { alert } from '@/api';
import type { ReceiverReceiver } from '@/api/generated';
import AlertDrawer from '@/components/AlertDrawer';
import IconTip from '@/components/IconTip';
import { Alert } from '@/type/alert';
import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import type { DrawerProps } from 'antd';
import { Button, Form, Input, Select, Space, message } from 'antd';
import { useEffect } from 'react';

/**
 * ChannelDrawer has three states: creation,editing and pure display
 * Editing state and pure display state can be switched between each other
 */
interface ChannelDrawerProps extends DrawerProps {
  status: Alert.DrawerStatus;
  name?: string;
  setStatus?: (status: Alert.DrawerStatus) => void;
  onClose: () => void;
  submitCallback?: () => void;
}

const { TextArea } = Input;

export default function ChannelDrawer({
  status,
  name,
  onClose,
  setStatus,
  submitCallback,
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
  const receiverNames = listReceiversRes?.data?.map(
    (receiver) => receiver.name,
  );
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
            {status === 'display'
              ? intl.formatMessage({
                  id: 'src.pages.Alert.Channel.D7E7D32B',
                  defaultMessage: '编辑',
                })
              : intl.formatMessage({
                  id: 'src.pages.Alert.Channel.A787FE28',
                  defaultMessage: '提交',
                })}
          </Button>
          <Button onClick={onClose}>
            {intl.formatMessage({
              id: 'src.pages.Alert.Channel.FDBFEACE',
              defaultMessage: '取消',
            })}
          </Button>
        </Space>
      </div>
    );
  };
  const submit = (values: ReceiverReceiver) => {
    alert.createOrUpdateReceiver(values).then(({ successful }) => {
      if (successful) {
        message.success(
          intl.formatMessage({
            id: 'src.pages.Alert.Channel.9091A7E9',
            defaultMessage: '操作成功!',
          }),
        );
        submitCallback && submitCallback();
        onClose();
      }
    });
  };
  const typeChange = (type: string) => {
    const template = listReceiverTemplates?.find((item) => item.type === type);
    if (template) form.setFieldValue('config', template.template);
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
      title={intl.formatMessage({
        id: 'src.pages.Alert.Channel.34E59D42',
        defaultMessage: '告警通道配置',
      })}
      footer={<Footer />}
      destroyOnClose={true}
      onSubmit={() => form.submit()}
      onClose={() => onClose()}
      {...props}
    >
      <Form form={form} onFinish={submit} preserve={false} layout="vertical">
        <Form.Item
          wrapperCol={{ span: 12 }}
          label={intl.formatMessage({
            id: 'src.pages.Alert.Channel.897D04CA',
            defaultMessage: '通道名称',
          })}
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.pages.Alert.Channel.6B0B7626',
                defaultMessage: '请输入',
              }),
            },
            {
              validator: (_, value) => {
                if (
                  status === 'create' &&
                  receiverNames?.some((receiver) => receiver === value)
                ) {
                  return Promise.reject(
                    intl.formatMessage({
                      id: 'src.pages.Alert.Channel.7965888F',
                      defaultMessage: '告警通道已存在，请重新输入',
                    }),
                  );
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
            <Input
              disabled={status !== 'create'}
              placeholder={intl.formatMessage({
                id: 'src.pages.Alert.Channel.DFFB3F07',
                defaultMessage: '请输入',
              })}
            />
          )}
        </Form.Item>
        <Form.Item
          wrapperCol={{ span: 12 }}
          label={intl.formatMessage({
            id: 'src.pages.Alert.Channel.B1B680BD',
            defaultMessage: '通道类型',
          })}
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.pages.Alert.Channel.A3D31BA5',
                defaultMessage: '请选择',
              }),
            },
          ]}
          name={'type'}
        >
          {status === 'display' ? (
            <p>{form.getFieldValue('type') || '-'} </p>
          ) : (
            <Select
              placeholder={intl.formatMessage({
                id: 'src.pages.Alert.Channel.AFFFAE43',
                defaultMessage: '请选择',
              })}
              onChange={typeChange}
              showSearch
              options={listReceiverTemplates?.map((template) => ({
                value: template.type,
                label: template.type,
              }))}
            />
          )}
        </Form.Item>
        <Form.Item
          name={'config'}
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'src.pages.Alert.Channel.BD932A40',
                defaultMessage: '请输入',
              }),
            },
          ]}
          label={
            <IconTip
              tip={
                <span>
                  {intl.formatMessage({
                    id: 'src.pages.Alert.Channel.5A64A44C',
                    defaultMessage:
                      '使用 yaml 格式配置告警通道，具体通道配置可参考',
                  })}{' '}
                  <a
                    href="https://prometheus.io/docs/alerting/latest/configuration/#receiver-integration-settings"
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    {intl.formatMessage({
                      id: 'src.pages.Alert.Channel.7F8C8DED',
                      defaultMessage: 'alertmanager 文档',
                    })}
                  </a>
                </span>
              }
              content={intl.formatMessage({
                id: 'src.pages.Alert.Channel.217A737A',
                defaultMessage: '通道配置',
              })}
            />
          }
        >
          {status === 'display' ? (
            <pre>{form.getFieldValue('config') || '-'}</pre>
          ) : (
            <TextArea
              rows={18}
              placeholder={intl.formatMessage({
                id: 'src.pages.Alert.Channel.6B980014',
                defaultMessage: '请输入',
              })}
            />
          )}
        </Form.Item>
      </Form>
    </AlertDrawer>
  );
}
