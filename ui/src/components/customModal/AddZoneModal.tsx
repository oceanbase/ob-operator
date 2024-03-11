import { intl } from '@/utils/intl';
import { Form, Input, InputNumber, message } from 'antd';

import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { addObzone } from '@/services';
import type { CommonModalType } from '.';
import CustomModal from '.';
import NodeSelector from '../NodeSelector';

type FieldType = {
  name: string;
};

export default function AddZoneModal({
  visible,
  setVisible,
  successCallback,
}: CommonModalType) {
  const [form] = Form.useForm();

  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };

  const handleCancel = () => setVisible(false);
  const onFinish = async (values: any) => {
    const [namespace, name] = getNSName();
    const res = await addObzone({ namespace, name, ...values });
    if (res.successful) {
      message.success(res.message);
      successCallback();
      form.resetFields();
      setVisible(false);
    }
  };

  const getNowNodeSelector = () => {
    return form.getFieldValue('nodeSelector');
  };
  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'OBDashboard.components.customModal.AddZoneModal.AddZone',
        defaultMessage: '新增Zone',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <Form
        form={form}
        name="dynamic_form_nest_item"
        onFinish={onFinish}
        style={{ maxWidth: 600 }}
        autoComplete="off"
      >
        <Form.Item<FieldType>
          label={intl.formatMessage({
            id: 'OBDashboard.components.customModal.AddZoneModal.ZoneName',
            defaultMessage: 'zone名称',
          })}
          name="zone"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'OBDashboard.components.customModal.AddZoneModal.EnterAZoneName',
                defaultMessage: '请输入zone名称!',
              }),
            },
          ]}
        >
          <Input
            placeholder={intl.formatMessage({
              id: 'OBDashboard.components.customModal.AddZoneModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
        <Form.Item<FieldType>
          label={intl.formatMessage({
            id: 'OBDashboard.components.customModal.AddZoneModal.NumberOfServers',
            defaultMessage: 'server数',
          })}
          name="replicas"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'OBDashboard.components.customModal.AddZoneModal.PleaseEnterTheNumberOf',
                defaultMessage: '请输入server数!',
              }),
            },
          ]}
        >
          <InputNumber
            placeholder={intl.formatMessage({
              id: 'OBDashboard.components.customModal.AddZoneModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
        {/* 如果不能为空的话应该需要添加默认值 */}
        {/* <p>node-selector:</p>
           <Form.List name="nodeSelector">
            {(fields, { add, remove }) => (
              <>
                {fields.map(({ key, name, ...restField }) => (
                  <Space
                    key={key}
                    style={{ display: 'flex', marginBottom: 8 }}
                    align="baseline"
                  >
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[{ required: true, message: '请输入key' }]}
                    >
                      <Input placeholder="key" />
                    </Form.Item>
                    :
                    <Form.Item
                      {...restField}
                      name={[name, 'value']}
                      rules={[{ required: true, message: '请输入value' }]}
                    >
                      <Input placeholder="value" />
                    </Form.Item>
                    <MinusCircleOutlined onClick={() => remove(name)} />
                  </Space>
                ))}
                <Form.Item>
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    block
                    icon={<PlusOutlined />}
                  >
                    添加node-selector
                  </Button>
                </Form.Item>
              </>
            )}
           </Form.List> */}
        <NodeSelector
          showLabel={true}
          formName="nodeSelector"
          getNowNodeSelector={getNowNodeSelector}
        />
      </Form>
    </CustomModal>
  );
}
