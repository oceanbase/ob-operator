import { intl } from '@/utils/intl';
import { Form, Input, message } from 'antd';

import { RULER_ZONE } from '@/constants/rules';
import { addObzoneReportWrap } from '@/services/reportRequest/clusterReportReq';
import { useParams } from '@umijs/max';
import type { CommonModalType } from '.';
import CustomModal from '.';
import InputNumber from '../InputNumber';
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
  const { ns: namespace, name } = useParams();
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };

  const handleCancel = () => {
    form.resetFields();
    setVisible(false);
  };
  const onFinish = async (values: any) => {
    const res = await addObzoneReportWrap({ namespace, name, ...values });
    if (res.successful) {
      message.success(
        res.message ||
          intl.formatMessage({
            id: 'Dashboard.components.customModal.AddZoneModal.OperationSucceeded',
            defaultMessage: '操作成功！',
          }),
      );
      if (successCallback) successCallback();
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
        initialValues={{ replicas: 1 }}
        style={{ maxWidth: 600 }}
        autoComplete="off"
      >
        <Form.Item<FieldType>
          label={intl.formatMessage({
            id: 'OBDashboard.components.customModal.AddZoneModal.ZoneName',
            defaultMessage: 'zone名称',
          })}
          name="zone"
          rules={RULER_ZONE}
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
