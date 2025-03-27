import { K8sClusterApi } from '@/api';
import { useRequest } from 'ahooks';
import { Form, Input, Modal, message } from 'antd';
import { isEmpty } from 'lodash';
import { useEffect } from 'react';

export default function Createk8sClusterModal({
  visible,
  editData,
  onSuccess,
  onCancel,
}: API.CommonModalType) {
  const [form] = Form.useForm<API.CreateClusterData>();
  const { resetFields, validateFields, setFieldsValue } = form;

  const isEdit = !isEmpty(editData);
  const { run: createK8sCluster, loading } = useRequest(
    K8sClusterApi.createRemoteK8sCluster,
    {
      manual: true,
      onSuccess: ({ successful }) => {
        if (successful) {
          message.success('创建 k8s 集群成功');
          onSuccess();
          resetFields();
          setFieldsValue({
            name: '',
            description: '',
            kubeConfig: '',
          });
        }
      },
    },
  );
  const { run: patchK8sCluster, loading: editLoading } = useRequest(
    K8sClusterApi.patchRemoteK8sCluster,
    {
      manual: true,
      onSuccess: ({ successful }) => {
        if (successful) {
          message.success('编辑 k8s 集群成功');
          onSuccess();
          resetFields();
          setFieldsValue({
            name: '',
            description: '',
            kubeConfig: '',
          });
        }
      },
    },
  );

  const handleSubmit = () => {
    validateFields().then((values) => {
      if (isEdit && !isEmpty(editDasta)) {
        patchK8sCluster(editData?.name, values);
      } else {
        createK8sCluster(values);
      }
    });
  };

  useEffect(() => {
    if (isEdit) {
      setFieldsValue({ ...editData });
    } else {
      setFieldsValue({
        name: '',
        description: '',
        kubeConfig: '',
      });
    }
  }, [visible]);

  return (
    <Modal
      title={isEdit ? '编辑 k8s 集群' : '创建 k8s 集群'}
      width={520}
      open={visible}
      onOk={() => handleSubmit()}
      onCancel={() => {
        onCancel();
        resetFields();
      }}
      confirmLoading={isEdit ? editLoading : loading}
    >
      <Form form={form} layout="vertical">
        <Form.Item
          name={'name'}
          label={'名称'}
          rules={[{ required: true, message: '请输入名称' }]}
        >
          <Input placeholder="请输入" disabled={isEdit} />
        </Form.Item>

        <Form.Item
          name={'description'}
          label={'描述信息'}
          rules={[{ required: true, message: '请输入描述信息' }]}
        >
          <Input placeholder="请输入" />
        </Form.Item>
        <Form.Item
          name={'kubeConfig'}
          label={'kubeConfig'}
          rules={[{ required: true, message: '请输入 kubeConfig' }]}
        >
          <Input.TextArea
            autoSize={{ minRows: 8, maxRows: 8 }}
            placeholder="请输入"
          />
        </Form.Item>
      </Form>
    </Modal>
  );
}
