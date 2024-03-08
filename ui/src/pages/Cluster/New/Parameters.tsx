import CollapsibleCard from '@/components/CollapsibleCard';
import { intl } from '@/utils/intl';
import { PlusOutlined } from '@ant-design/icons';
import { Button, Col, Form, Input, Popconfirm, Row } from 'antd';

export default function Parameters() {
  return (
    <Col span={24}>
      <CollapsibleCard
        title={intl.formatMessage({
          id: 'Dashboard.Cluster.New.Parameters.ParameterSettings',
          defaultMessage: '参数设置',
        })}
        collapsible={true}
        bordered={false}
      >
        <Form.List name="parameters">
          {(fields, { add, remove }) => (
            <>
              {fields.map(({ key, name }, index) => (
                <div key={key}>
                  <Row gutter={8}>
                    <Col span={6}>
                      <Form.Item
                        label={
                          index === 0 &&
                          intl.formatMessage({
                            id: 'OBDashboard.Cluster.New.Parameters.ParameterName',
                            defaultMessage: '参数名',
                          })
                        }
                        name={[name, 'key']}
                        rules={[
                          {
                            required: true,
                            message: intl.formatMessage({
                              id: 'OBDashboard.Cluster.New.Parameters.EnterAParameterName',
                              defaultMessage: '请输入参数名',
                            }),
                          },
                        ]}
                      >
                        <Input
                          placeholder={intl.formatMessage({
                            id: 'OBDashboard.Cluster.New.Parameters.PleaseEnter',
                            defaultMessage: '请输入',
                          })}
                        />
                      </Form.Item>
                    </Col>
                    <Col span={6}>
                      <Form.Item
                        label={
                          index === 0 &&
                          intl.formatMessage({
                            id: 'OBDashboard.Cluster.New.Parameters.ParameterValue',
                            defaultMessage: '参数值',
                          })
                        }
                        name={[name, 'value']}
                        rules={[
                          {
                            required: true,
                            message: intl.formatMessage({
                              id: 'OBDashboard.Cluster.New.Parameters.EnterAParameterValue',
                              defaultMessage: '请输入参数值',
                            }),
                          },
                        ]}
                      >
                        <Input
                          placeholder={intl.formatMessage({
                            id: 'OBDashboard.Cluster.New.Parameters.PleaseEnter',
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
                            id: 'OBDashboard.Cluster.New.Parameters.Operation',
                            defaultMessage: '操作',
                          })
                        }
                      >
                        <Popconfirm
                          placement="left"
                          title={intl.formatMessage({
                            id: 'OBDashboard.Cluster.New.Parameters.AreYouSureYouWant',
                            defaultMessage: '确定要删除该参数吗？',
                          })}
                          onConfirm={() => {
                            remove(index);
                          }}
                          okText={intl.formatMessage({
                            id: 'OBDashboard.Cluster.New.Parameters.Delete',
                            defaultMessage: '删除',
                          })}
                          cancelText={intl.formatMessage({
                            id: 'OBDashboard.Cluster.New.Parameters.Cancel',
                            defaultMessage: '取消',
                          })}
                          okButtonProps={{
                            danger: true,
                            ghost: true,
                          }}
                        >
                          <a>
                            {intl.formatMessage({
                              id: 'OBDashboard.Cluster.New.Parameters.Delete',
                              defaultMessage: '删除',
                            })}
                          </a>
                        </Popconfirm>
                      </Form.Item>
                    </Col>
                  </Row>
                </div>
              ))}
              <Col span={13}>
                <Form.Item>
                  <Button
                    type="dashed"
                    onClick={() => add()}
                    block
                    icon={<PlusOutlined />}
                  >
                    {intl.formatMessage({
                      id: 'OBDashboard.Cluster.New.Parameters.AddParameters',
                      defaultMessage: '添加参数',
                    })}
                  </Button>
                </Form.Item>
              </Col>
            </>
          )}
        </Form.List>
      </CollapsibleCard>
    </Col>
  );
}
