import { intl } from '@/utils/intl';
import { DeleteOutlined,PlusOutlined } from '@ant-design/icons';
import type { FormInstance } from 'antd';
import {
Button,
Card,
Col,
Form,
Input,
InputNumber,
Popconfirm,
Row,
} from 'antd';

import NodeSelector from '@/components/NodeSelector';
import { TZ_NAME_REG } from '@/constants';
import { resourceNameRule } from './helper';

export default function Topo({ form }: { form: FormInstance<any> }) {
  const getNowNodeSelector = (zoneIdx: number) => {
    return () => {
      const topologyData = form.getFieldValue('topology');
      const { nodeSelector } = topologyData[zoneIdx];
      return nodeSelector;
    };
  };

  return (
    <Col span={24}>
      <Card
        title={intl.formatMessage({
          id: 'dashboard.Cluster.New.Topo.Topology',
          defaultMessage: '拓扑',
        })}
      >
        <Form.List name="topology">
          {(fields, { add, remove }) => {
            return (
              <>
                {fields.map(({ key, name }, index) => (
                  <div key={key}>
                    <Row gutter={8}>
                      <Col span={6}>
                        <Form.Item
                          label={
                            index === 0 &&
                            intl.formatMessage({
                              id: 'OBDashboard.Cluster.New.Topo.ZoneName',
                              defaultMessage: 'Zone名称',
                            })
                          }
                          validateFirst
                          name={[name, 'zone']}
                          rules={[
                            {
                              required: true,
                              message: intl.formatMessage({
                                id: 'OBDashboard.Cluster.New.Topo.EnterAZoneName',
                                defaultMessage: '请输入zone名称',
                              }),
                            },
                            {
                              pattern: TZ_NAME_REG,
                              message: '首字符必须是字母或者下划线，不能包含 -',
                            },
                            resourceNameRule,
                          ]}
                        >
                          <Input
                            placeholder={intl.formatMessage({
                              id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                              defaultMessage: '请输入',
                            })}
                          />
                        </Form.Item>
                      </Col>
                      <Col span={12}>
                        <NodeSelector
                          formName={[name, 'nodeSelector']}
                          showLabel={index === 0}
                          getNowNodeSelector={getNowNodeSelector(index)}
                        />
                      </Col>
                      <Col>
                        <Form.Item
                          label={
                            index === 0 &&
                            intl.formatMessage({
                              id: 'OBDashboard.Cluster.New.Topo.NumberOfServers',
                              defaultMessage: 'server数',
                            })
                          }
                          name={[name, 'replicas']}
                        >
                          <InputNumber
                            placeholder={intl.formatMessage({
                              id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                              defaultMessage: '请输入',
                            })}
                            min={1}
                          />
                        </Form.Item>
                      </Col>
                      <Col span={1}>
                        {fields.length > 1 && (
                          <Form.Item
                            label={index === 0 && ' '}
                            style={{ marginBottom: 8 }}
                            name={[name, ' ']}
                          >
                            <Popconfirm
                              placement="left"
                              title={intl.formatMessage({
                                id: 'OBDashboard.Cluster.New.Topo.AreYouSureYouWant',
                                defaultMessage: '确定要删除该配置项吗？',
                              })}
                              onConfirm={() => {
                                remove(index);
                              }}
                              okText={intl.formatMessage({
                                id: 'OBDashboard.Cluster.New.Topo.Delete',
                                defaultMessage: '删除',
                              })}
                              cancelText={intl.formatMessage({
                                id: 'OBDashboard.Cluster.New.Topo.Cancel',
                                defaultMessage: '取消',
                              })}
                              okButtonProps={{
                                danger: true,
                                ghost: true,
                              }}
                            >
                              <DeleteOutlined
                                style={{ color: 'rgba(0, 0, 0, .45)' }}
                              />
                            </Popconfirm>
                          </Form.Item>
                        )}
                      </Col>
                    </Row>
                  </div>
                ))}
                <Row gutter={8}>
                  <Col span={20}>
                    <Form.Item>
                      <Button
                        type="dashed"
                        onClick={() =>
                          add({
                            zone: `zone${fields.length + 1}`,
                            nodeSelector: [],
                            replicas: 1,
                          })
                        }
                        block
                        icon={<PlusOutlined />}
                      >
                        {intl.formatMessage({
                          id: 'OBDashboard.Cluster.New.Topo.AddZone',
                          defaultMessage: '添加Zone',
                        })}
                      </Button>
                    </Form.Item>
                  </Col>
                </Row>
              </>
            );
          }}
        </Form.List>
      </Card>
    </Col>
  );
}
