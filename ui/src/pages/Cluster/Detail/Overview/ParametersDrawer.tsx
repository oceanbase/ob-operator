import { obproxy } from '@/api';
import { intl } from '@/utils/intl';
import { PlusOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import {
  Button,
  Col,
  Drawer,
  Form,
  Input,
  Popconfirm,
  Row,
  Space,
  message,
} from 'antd';
import React from 'react';

export interface ParametersModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  initialValues: any[];
}

const ParametersDrawer: React.FC<ParametersModalProps> = ({
  visible,
  onCancel,
  initialValues,
  onSuccess,
}) => {
  const [form] = Form.useForm<API.CreateClusterData>();
  const { validateFields } = form;

  const { runAsync: updateParameters, loading } = useRequest(
    obproxy.patchOBProxy,
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
      title={intl.formatMessage({
        id: 'src.pages.Cluster.Detail.Overview.849E8956',
        defaultMessage: '参数编辑',
      })}
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
              validateFields().then((values) => {
                const { parameters } = values;
                updateParameters({
                  parameters,
                });
              });
            }}
          >
            确定
          </Button>
        </Space>
      }
    >
      <Form form={form} layout="vertical" style={{ marginBottom: 56 }}>
        <Form.List name="parameters" initialValue={initialValues}>
          {(fields, { add, remove }) => (
            <>
              {fields.map(({ key, name }, index) => (
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
                    <Col>
                      <Form.Item
                        label={
                          index === 0 &&
                          intl.formatMessage({
                            id: 'src.pages.Cluster.Detail.Overview.C52AC4A3',
                            defaultMessage: '操作',
                          })
                        }
                      >
                        <Popconfirm
                          placement="left"
                          title={intl.formatMessage({
                            id: 'src.pages.Cluster.Detail.Overview.8B44D940',
                            defaultMessage: '确定要删除该参数吗？',
                          })}
                          onConfirm={() => {
                            remove(index);
                          }}
                          okText={intl.formatMessage({
                            id: 'src.pages.Cluster.Detail.Overview.3D14CA3E',
                            defaultMessage: '删除',
                          })}
                          cancelText={intl.formatMessage({
                            id: 'src.pages.Cluster.Detail.Overview.476938E4',
                            defaultMessage: '取消',
                          })}
                          okButtonProps={{
                            danger: true,
                            ghost: true,
                          }}
                        >
                          <a style={{ color: 'red' }}>
                            {intl.formatMessage({
                              id: 'src.pages.Cluster.Detail.Overview.86ECCAB9',
                              defaultMessage: '删除',
                            })}
                          </a>
                        </Popconfirm>
                      </Form.Item>
                    </Col>
                  </Row>
                </div>
              ))}
              <Col span={20}>
                <Form.Item>
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    block
                    icon={<PlusOutlined />}
                  >
                    {intl.formatMessage({
                      id: 'src.pages.Cluster.Detail.Overview.23FEC744',
                      defaultMessage: '添加参数',
                    })}
                  </Button>
                </Form.Item>
              </Col>
            </>
          )}
        </Form.List>
      </Form>
    </Drawer>
  );
};

export default ParametersDrawer;
