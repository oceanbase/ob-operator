import { SUFFIX_UNIT } from '@/constants';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { patchTenantConfiguration } from '@/services/tenant';
import { intl } from '@/utils/intl';
import { formatUnitDetailData } from '@/utils/helper';

import { Col,Form,InputNumber,Row,message } from 'antd';
import type { CommonModalType } from '.';
import CustomModal from '.';

export type UnitDetailType = {
  cpuCount: number;
  iopsWeight: number;
  logDiskSize: number;
  maxIops: number;
  memorySize: number;
  minIops: number;
};

export default function ModifyUnitDetailModal({
  visible,
  setVisible,
  successCallback,
}: CommonModalType) {
  const [form] = Form.useForm<UnitDetailType>();
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };
  const handleCancel = () => setVisible(false);
  const onFinish = async (values: any) => {
    const [ns, name] = getNSName();
    const res = await patchTenantConfiguration({ ns, name, ...formatUnitDetailData(values) });
    if (res.successful) {
      message.success(res.message);
      successCallback();
      form.resetFields();
      setVisible(false);
    }
  };
  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'Dashboard.components.customModal.ModifyUnitDetailModal.AdjustUnitSpecifications',
        defaultMessage: '调整 Unit 规格',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <Form
        form={form}
        onFinish={onFinish}
        style={{ maxWidth: 600 }}
        autoComplete="off"
      >
        <Form.Item
          label="CPU"
          name="cpuCount"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EnterTheNumberOfCpu',
                defaultMessage: '请输入 CPU 核数',
              }),
            },
          ]}
        >
          <InputNumber
            addonAfter={intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyUnitDetailModal.Nuclear',
              defaultMessage: '核',
            })}
            placeholder={intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
        <Form.Item
          label="Memory"
          name="memorySize"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EnterMemory',
                defaultMessage: '请输入 Memory',
              }),
            },
          ]}
        >
          <InputNumber
            addonAfter={SUFFIX_UNIT}
            placeholder={intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
        <Form.Item label="LogDiskSize" name="logDiskSize">
          <InputNumber
            addonAfter={SUFFIX_UNIT}
            placeholder={intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
        <Row gutter={24}>
          <Col>
            <Form.Item label="min iops" name="minIops">
              <InputNumber
                placeholder={intl.formatMessage({
                  id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col>
            <Form.Item label="max iops" name="maxIops">
              <InputNumber
                placeholder={intl.formatMessage({
                  id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
        </Row>
        <Form.Item
          label={intl.formatMessage({
            id: 'Dashboard.components.customModal.ModifyUnitDetailModal.IopsWeight',
            defaultMessage: 'iops权重',
          })}
          name="iopsWeight"
        >
          <InputNumber
            placeholder={intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
