import { intl } from '@/utils/intl';
import { Button, Col, Form, Input, Modal, Row, Space } from 'antd';
import React from 'react';

export interface ParametersModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  initialValues: any[];
  name: string;
  namespace: string;
}

const ParametersModal: React.FC<ParametersModalProps> = ({
  visible,
  onCancel,
  initialValues,
  // onSuccess,
  // name,
  // namespace,
}) => {
  const [form] = Form.useForm<API.CreateClusterData>();
  const { validateFields } = form;

  // const { runAsync: updateParameters, loading } = useRequest(
  //   obcluster.patchOBCluster,
  //   {
  //     manual: true,
  //     onSuccess: (res) => {
  //       if (res.successful) {
  //         message.success(
  //           intl.formatMessage({
  //             id: 'src.pages.Cluster.Detail.Overview.E908AA54',
  //             defaultMessage: '编辑参数已成功',
  //           }),
  //         );
  //         onSuccess();
  //       }
  //     },
  //   },
  // );

  return (
    <Modal
      title={intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.849E8956',
        defaultMessage: '参数编辑',
      })}
      open={visible}
      destroyOnClose
      onCancel={() => onCancel()}
      width={520}
      footer={
        <Space>
          <Button onClick={onCancel}>取消</Button>
          <Button
            type="primary"
            // loading={loading}
            // TODO 单独编辑参数接口
            onClick={() => {
              validateFields().then((values) => {
                console.log('values', values);
                // updateParameters(name, namespace, values, `编辑参数已成功`);
              });
            }}
          >
            确定
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
              initialValue={initialValues?.name}
              name={'name'}
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
              initialValue={initialValues?.value}
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
