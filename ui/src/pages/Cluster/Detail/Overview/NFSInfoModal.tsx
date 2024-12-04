import { obcluster } from '@/api';
import { useRequest } from 'ahooks';
import { Form, Input, message, Modal } from 'antd';
import React from 'react';

export interface NFSInfoModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  removeNFS: boolean;
  title: string;
  name: string;
  namespace: string;
}

const NFSInfoModal: React.FC<NFSInfoModalProps> = ({
  visible,
  onCancel,
  removeNFS,
  onSuccess,
  title,
  name,
  namespace,
}) => {
  const [form] = Form.useForm<FormData>();

  const { runAsync: patchOBCluster, loading } = useRequest(
    obcluster.patchOBCluster,
    {
      manual: true,
      onSuccess: (res) => {
        if (res.successful) {
          message.success(`修改${title}成功`);
          onSuccess();
        }
      },
    },
  );

  return (
    <Modal
      title={title}
      open={visible}
      destroyOnClose
      onCancel={() => onCancel()}
      confirmLoading={loading}
      onOk={() => {
        if (removeNFS) {
          const body = {
            removeBackupVolume: true,
          };
          patchOBCluster(name, namespace, body);
        } else {
          form.validateFields().then((values) => {
            const { address, path } = values;
            const body = {
              backupVolume: {
                address,
                path,
              },
            };
            patchOBCluster(name, namespace, body);
          });
        }
      }}
    >
      {removeNFS ? (
        ' 注意，移除挂载的 NFS 备份卷会滚动重启所有节点，确认移除吗？'
      ) : (
        <Form form={form}>
          <Form.Item
            label="地址"
            name="address"
            rules={[
              {
                required: true,
                message: '请输入地址',
              },
            ]}
          >
            <Input placeholder={'请输入'} />
          </Form.Item>
          <Form.Item
            label="路径"
            name="path"
            rules={[
              {
                required: true,
                message: '请输入路径',
              },
            ]}
          >
            <Input placeholder={'请输入'} />
          </Form.Item>
          <Form.Item noStyle>注意，挂载备份卷会滚动重启所有节点</Form.Item>
        </Form>
      )}
    </Modal>
  );
};

export default NFSInfoModal;
