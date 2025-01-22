import { obcluster } from '@/api';
import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Button, Col, Form, Input, message, Modal, Row, Space } from 'antd';
import React, { useEffect } from 'react';

interface kvPair {
  key: string;
  value: string;
}

export interface ParametersModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  initialValues: kvPair;
  name: string;
  namespace: string;
}

const ParametersModal: React.FC<ParametersModalProps> = ({
  visible,
  onCancel,
  initialValues,
  onSuccess,
  name,
  namespace,
}) => {
  const [form] = Form.useForm<kvPair>();
  const { validateFields, resetFields } = form;

  useEffect(() => {
    if (visible) {
      form.setFieldsValue(initialValues);
    }
  }, [visible]);

  const { runAsync: updateParameters, loading } = useRequest(
    obcluster.patchOBCluster,
    {
      manual: true,
      onSuccess: (res) => {
        if (res.successful) {
          message.success(
            intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.E908AA54',
              defaultMessage: '编辑参数已成功',
            }),
          );
          onSuccess();
        }
      },
    },
  );

  return (
    <Modal
      title={intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.849E8956',
        defaultMessage: '参数编辑',
      })}
      maskClosable={false}
      open={visible}
      onCancel={() => {
        onCancel();
      }}
      width={520}
      footer={
        <Space>
          <Button
            onClick={() => {
              onCancel();
              resetFields();
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.42228E22',
              defaultMessage: '取消',
            })}
          </Button>
          <Button
            type="primary"
            loading={loading}
            onClick={() => {
              validateFields().then((values) => {
                const objValue = {
                  modifiedParameters: [values],
                };
                updateParameters(
                  namespace,
                  name,
                  objValue,
                  intl.formatMessage({
                    id: 'src.pages.Cluster.Detail.Overview.480E47F8',
                    defaultMessage: '编辑参数已成功',
                  }),
                );
              });
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.9F17E32D',
              defaultMessage: '确定',
            })}
          </Button>
        </Space>
      }
    >
      <Form form={form} layout="vertical" style={{ marginBottom: 56 }}>
        <Row gutter={4}>
          <Col span={12}>
            <Form.Item
              label={intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.0F9AD89D',
                defaultMessage: '参数名',
              })}
              name={'key'}
            >
              <Input disabled={true} />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              label={intl.formatMessage({
                id: 'src.pages.Cluster.Detail.Overview.4E02366B',
                defaultMessage: '参数值',
              })}
              name={'value'}
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'src.pages.Cluster.Detail.Overview.5DA70D14',
                    defaultMessage: '请输入参数值',
                  }),
                },
              ]}
            >
              <Input
                placeholder={intl.formatMessage({
                  id: 'src.pages.Cluster.Detail.Overview.E26B7DFD',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </Modal>
  );
};

export default ParametersModal;
