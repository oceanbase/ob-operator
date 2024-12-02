import { obcluster } from '@/api';
import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Button, Col, Drawer, Form, Input, message, Row, Space } from 'antd';
import React from 'react';

export interface ParametersModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  initialValues: any[];
  name: string;
  namespace: string;
}

const ResourceDrawer: React.FC<ParametersModalProps> = ({
  visible,
  onCancel,
  initialValues,
  name,
  namespace,
  onSuccess,
}) => {
  const [form] = Form.useForm<API.CreateClusterData>();
  const { validateFields } = form;

  const { runAsync: patchOBCluster, loading } = useRequest(
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
    <Drawer
      title={'存储资源编辑'}
      open={visible}
      destroyOnClose
      onClose={() => onCancel()}
      width={520}
      footer={
        <Space>
          <Button onClick={onCancel}>取消</Button>
          <Button
            type="primary"
            loading={loading}
            onClick={() => {
              validateFields().then((value) => {
                const { resource } = value;

                const date = resource
                  ?.filter((item) => item.type === 'data')
                  ?.map((item1) =>
                    item1.label === 'size'
                      ? { size: item1.value }
                      : { storageClass: item1.value },
                  );
                const log = resource
                  ?.filter((item) => item.type === 'log')
                  ?.map((item1) =>
                    item1.label === 'size'
                      ? { size: item1.value }
                      : { storageClass: item1.value },
                  );
                const redoLog = resource
                  ?.filter((item) => item.type === 'redoLog')
                  ?.map((item1) =>
                    item1.label === 'size'
                      ? { size: item1.value }
                      : { storageClass: item1.value },
                  );

                const body = {
                  storage: {
                    data: date,
                    log: log,
                    redoLog: redoLog,
                  },
                };
                patchOBCluster(name, namespace, body, `存储资源编辑成功`);
              });
            }}
          >
            确定
          </Button>
        </Space>
      }
    >
      <Form form={form} layout="vertical" style={{ marginBottom: 56 }}>
        <Form.List name="resource" initialValue={initialValues}>
          {(fields) => (
            <>
              {fields.map(({ key, name }, index) => {
                return (
                  <div key={key}>
                    <Row gutter={8}>
                      <Col span={10}>
                        <Form.Item
                          label={
                            index === 0 &&
                            intl.formatMessage({
                              id: 'src.pages.Cluster.Detail.Overview.0F9AD89D',
                              defaultMessage: '参数名',
                            })
                          }
                          name={[name, 'key']}
                          rules={[
                            {
                              required: true,
                              message: intl.formatMessage({
                                id: 'src.pages.Cluster.Detail.Overview.F0473B44',
                                defaultMessage: '请输入参数名',
                              }),
                            },
                          ]}
                        >
                          <Input
                            placeholder={intl.formatMessage({
                              id: 'src.pages.Cluster.Detail.Overview.C118F812',
                              defaultMessage: '请输入',
                            })}
                          />
                        </Form.Item>
                      </Col>
                      <Col span={10}>
                        <Form.Item
                          label={
                            index === 0 &&
                            intl.formatMessage({
                              id: 'src.pages.Cluster.Detail.Overview.4E02366B',
                              defaultMessage: '参数值',
                            })
                          }
                          name={[name, 'value']}
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
                  </div>
                );
              })}
            </>
          )}
        </Form.List>
      </Form>
    </Drawer>
  );
};

export default ResourceDrawer;
