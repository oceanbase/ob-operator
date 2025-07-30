import { obcluster } from '@/api';
import { intl } from '@/utils/intl';
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

  const { validateFields, resetFields } = form;
  const { runAsync: patchOBCluster, loading } = useRequest(
    obcluster.patchOBCluster,
    {
      manual: true,
      onSuccess: (res) => {
        if (res.successful) {
          message.success(
            intl.formatMessage(
              {
                id: 'src.pages.Cluster.Detail.Overview.04490EC8',
                defaultMessage: '修改${title}成功',
              },
              { title: title },
            ),
          );
          onSuccess();
          resetFields();
        }
      },
    },
  );

  return (
    <Modal
      title={title}
      open={visible}
      destroyOnClose
      maskClosable={false}
      onCancel={() => {
        onCancel();
        resetFields();
      }}
      confirmLoading={loading}
      onOk={() => {
        if (removeNFS) {
          const body = {
            removeBackupVolume: true,
          };
          patchOBCluster(namespace, name, body);
        } else {
          validateFields().then((values) => {
            const { address, path } = values;
            const body = {
              backupVolume: {
                address,
                path,
              },
            };
            patchOBCluster(namespace, name, body);
          });
        }
      }}
    >
      {removeNFS ? (
        intl.formatMessage({
          id: 'src.pages.Cluster.Detail.Overview.426E8CD7',
          defaultMessage:
            ' 注意，移除挂载的 NFS 备份卷会滚动重启所有节点，确认移除吗？',
        })
      ) : (
        <Form form={form}>
          <Form.Item
            label={intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.DB5B21F0',
              defaultMessage: '地址',
            })}
            name="address"
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.60033114',
                  defaultMessage: '请输入地址',
                }),
              },
            ]}
          >
            <Input
              placeholder={intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.D16F4B6E',
                defaultMessage: '例如 172.17.x.x',
              })}
            />
          </Form.Item>
          <Form.Item
            label={intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.ACC053A9',
              defaultMessage: '路径',
            })}
            name="path"
            rules={[
              {
                required: true,
                message: intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.B7E9F065',
                  defaultMessage: '请输入路径',
                }),
              },
            ]}
          >
            <Input
              placeholder={intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.D0356ACB',
                defaultMessage: '例如 /opt/nfs',
              })}
            />
          </Form.Item>
          <Form.Item noStyle>
            {intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.0798E33D',
              defaultMessage: '注意，挂载备份卷会滚动重启所有节点',
            })}
          </Form.Item>
        </Form>
      )}
    </Modal>
  );
};

export default NFSInfoModal;
