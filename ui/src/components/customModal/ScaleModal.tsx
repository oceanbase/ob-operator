import { intl } from '@/utils/intl';
import { Form, InputNumber, message } from 'antd';
import CustomModal from '.';

import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { scaleObserverReportWrap } from '@/services/reportRequest/clusterReportReq';
import { useModel } from '@umijs/max';
import { useEffect } from 'react';
import type { CommonModalType } from '.';

interface ScaleModalProps {
  params?:{
    zoneName?: string;
    defaultValue?: number;
  }
}

export default function ScaleModal({
  visible,
  setVisible,
  successCallback,
  params,
}: ScaleModalProps & CommonModalType) {
  const { appInfo } = useModel('global');
  const zoneName = params?.zoneName;
  const defaultValue = params?.defaultValue;
  const [form] = Form.useForm();
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };
  const onFinish = async (val: any) => {
    if (!zoneName) throw new Error('zoneName is not defined');
    const [namespace, name] = getNSName();
    const res = await scaleObserverReportWrap({
      namespace,
      name,
      zoneName,
      replicas: val.replicas,
      version: appInfo.version
    });
    if (res.successful) {
      message.success(res.message);
      if(successCallback) successCallback();
      form.resetFields();
      setVisible(false);
    }
  };

  useEffect(() => {
    form.setFieldValue('replicas', defaultValue);
  }, [defaultValue]);

  return (
    <CustomModal
      title={intl.formatMessage({
        id: 'OBDashboard.components.customModal.ScaleModal.ExpandZone',
        defaultMessage: '扩容Zone',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={() => {
        setVisible(false);
        form.resetFields();
      }}
    >
      <Form form={form} onFinish={onFinish}>
        <Form.Item
          label="servers"
          name="replicas"
          rules={[
            {
              required: true,
              message: intl.formatMessage({
                id: 'OBDashboard.components.customModal.ScaleModal.PleaseEnterTheNumberOf',
                defaultMessage: '请输入扩缩容数目!',
              }),
            },
          ]}
        >
          <InputNumber
            defaultValue={defaultValue}
            min={1}
            placeholder={intl.formatMessage({
              id: 'OBDashboard.components.customModal.ScaleModal.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Form>
    </CustomModal>
  );
}
