import { K8sClusterApi } from '@/api';
import { encryptText, usePublicKey } from '@/hook/usePublicKey';
import { useRequest } from 'ahooks';
import { Form, Input, Modal, message } from 'antd';
import CryptoJS from 'crypto-js';
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

  function generateAESKey() {
    return CryptoJS.lib.WordArray.random(32).toString(CryptoJS.enc.Hex);
  }
  // Function to generate a random IV
  function generateIV() {
    return CryptoJS.lib.WordArray.random(16).toString(CryptoJS.enc.Hex);
  }

  // Function to encrypt data using AES-256
  function encryptAES(data, key, iv) {
    const encrypted = CryptoJS.AES.encrypt(data, CryptoJS.enc.Hex.parse(key), {
      iv: CryptoJS.enc.Hex.parse(iv),
      mode: CryptoJS.mode.CBC,
      padding: CryptoJS.pad.Pkcs7,
    });

    // Concatenate IV and ciphertext, then encode in base64
    const ivHex = CryptoJS.enc.Hex.parse(iv);
    const ciphertext = ivHex.clone().concat(encrypted.ciphertext);
    return ciphertext.toString(CryptoJS.enc.Base64);
  }

  const publicKey = usePublicKey();
  const handleSubmit = () => {
    validateFields().then((values) => {
      const key = generateAESKey();
      const iv = generateIV();
      const encryptedData = encryptAES(values.kubeConfig, key, iv);
      values.kubeConfig = encryptedData;

      const encryptedKey = encryptText(key, publicKey);

      if (isEdit && !isEmpty(editData)) {
        patchK8sCluster(editData?.name, values, {
          headers: {
            'X-Encrypted-Key': encryptedKey,
          },
        });
      } else {
        createK8sCluster(values, {
          headers: {
            'X-Encrypted-Key': encryptedKey,
          },
        });
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
