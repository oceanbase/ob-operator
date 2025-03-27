import { intl } from '@/utils/intl';
import { Form, Input, message, Select } from 'antd';

import { RULER_ZONE } from '@/constants/rules';
import { addObzoneReportWrap } from '@/services/reportRequest/clusterReportReq';
import { useParams } from '@umijs/max';

import { getK8sObclusterListReq } from '@/services';
import { useRequest } from 'ahooks';
import CustomModal from '.';
import InputNumber from '../InputNumber';
import NodeSelector from '../NodeSelector';

interface FormData {
  zone: string;
  replicas: number;
  nodeSelector: { key: string; value: string }[];
}

export default function AddZoneModal({
  visible,
  setVisible,
  successCallback,
}: API.CommonModalType) {
  const [form] = Form.useForm<FormData>();
  const { ns, name } = useParams();
  const { data: K8sClustersList } = useRequest(getK8sObclusterListReq);
  const options = (K8sClustersList?.data || [])?.map((item) => ({
    value: item.name,
    label: item.name,
  }));
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };

  const onCancel = () => {
    form.resetFields();
    setVisible(false);
  };
  const onFinish = async (values: FormData) => {
    const res = await addObzoneReportWrap({
      namespace: ns!,
      name: name!,
      ...values,
    });
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
      open={visible}
      onOk={handleSubmit}
      onCancel={onCancel}
    >
      <Form
        form={form}
        name="dynamic_form_nest_item"
        onFinish={onFinish}
        initialValues={{ replicas: 1 }}
        style={{ maxWidth: 600 }}
        autoComplete="off"
      >
        <Form.Item
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
        <Form.Item
          label={'K8s 集群'}
          name={'k8sCluster'}
          rules={[
            {
              required: true,
              message: '请选择 K8s 集群',
            },
          ]}
        >
          <Select
            showSearch
            placeholder="请选择 K8s 集群"
            optionFilterProp="label"
            options={options}
          />
        </Form.Item>
        <Form.Item
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
        <NodeSelector
          showLabel={true}
          formName="nodeSelector"
          getNowNodeSelector={getNowNodeSelector}
        />
      </Form>
    </CustomModal>
  );
}
