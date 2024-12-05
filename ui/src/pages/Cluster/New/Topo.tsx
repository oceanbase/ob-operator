import { intl } from '@/utils/intl';
import {
  DeleteOutlined,
  DownOutlined,
  PlusOutlined,
  RightOutlined,
} from '@ant-design/icons';
import {
  Button,
  Card,
  Col,
  Dropdown,
  Form,
  Input,
  InputNumber,
  Menu,
  Popconfirm,
  Row,
  Select,
  Space,
} from 'antd';

import { RULER_ZONE } from '@/constants/rules';
import { useState } from 'react';

export default function Topo() {
  // const { getFieldValue } = form;

  const [topologyConfiguration, setTopologyConfiguration] =
    useState<string>('');
  const [showTopology, setShowTopology] = useState<boolean>(false);

  // const getNowNodeSelector = (zoneIdx: number) => {
  //   return () => {
  //     const topologyData = getFieldValue('topology');
  //     const { nodeSelector } = topologyData[zoneIdx];
  //     return nodeSelector;
  //   };
  // };

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
                  <Card style={{ marginBottom: 16 }}>
                    <div key={key}>
                      <Row gutter={8}>
                        <Col span={12}>
                          <Form.Item
                            label={
                              // index === 0 &&
                              intl.formatMessage({
                                id: 'OBDashboard.Cluster.New.Topo.ZoneName',
                                defaultMessage: 'Zone名称',
                              })
                            }
                            validateFirst
                            name={[name, 'zone']}
                            rules={RULER_ZONE}
                          >
                            <Input
                              placeholder={intl.formatMessage({
                                id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                                defaultMessage: '请输入',
                              })}
                            />
                          </Form.Item>
                        </Col>
                        {/* <Col span={12}>
                          <NodeSelector
                            formName={[name, 'nodeSelector']}
                            showLabel={index === 0}
                            getNowNodeSelector={getNowNodeSelector(index)}
                          />
                        </Col> */}
                        <Col span={6}>
                          <Form.Item
                            label={
                              // index === 0 &&
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
                        <Col span={5}>
                          <Form.Item label={'Topology summary'}>-</Form.Item>
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
                        <Space
                          onClick={() => {
                            setShowTopology(!showTopology);
                          }}
                        >
                          {showTopology ? <DownOutlined /> : <RightOutlined />}
                          Topology
                        </Space>
                      </Row>
                      {/* 
                      TOdo 
                      表单存在两个问题：
                      1 点击新增 Topology 按钮，多个card 会同时展示下拉标志
                      2 点击新增 Topology 按钮，原有选择nodeselector 时，后选择pod affinity 时，pod affinity 会覆盖 nodeselector 的选择的UI
                       */}
                      {showTopology && (
                        <>
                          <div
                            style={{
                              paddingTop: '16px',
                              paddingBottom: '16px',
                            }}
                          >
                            {topologyConfiguration}
                          </div>
                          <Form.Item>
                            <Form.List name={[name, 'list']}>
                              {(subFields, subOpt) => (
                                <div
                                  style={{
                                    display: 'flex',
                                    flexDirection: 'column',
                                    rowGap: 16,
                                  }}
                                >
                                  {subFields.map((subField) => (
                                    <div key={subField.key}>
                                      <Row gutter={8}>
                                        {topologyConfiguration ===
                                          'Pod Affinity' && (
                                          <Col span={6}>
                                            <Form.Item
                                              label={'Type'}
                                              name={[name, 'type']}
                                            >
                                              <Select
                                                options={[
                                                  {
                                                    value: 'Affinity',
                                                    label: 'Affinity',
                                                  },
                                                  {
                                                    value: 'AntiAffinity',
                                                    label: 'AntiAffinity',
                                                  },
                                                ]}
                                              />
                                            </Form.Item>
                                          </Col>
                                        )}

                                        {topologyConfiguration ===
                                          'Toleration' && (
                                          <Col span={4}>
                                            <Form.Item
                                              label={'Effect'}
                                              name={[name, 'Effect']}
                                            >
                                              <Select
                                                options={[
                                                  {
                                                    value: 'NoSchedule',
                                                    label: 'NoSchedule',
                                                  },
                                                  {
                                                    value: 'PerferNoSchedule',
                                                    label: 'PerferNoSchedule',
                                                  },
                                                  {
                                                    value: 'NoExecute',
                                                    label: 'NoExecute',
                                                  },
                                                ]}
                                              />
                                            </Form.Item>
                                          </Col>
                                        )}
                                        <Col
                                          span={
                                            topologyConfiguration ===
                                            'nodeSelector'
                                              ? 7
                                              : 5
                                          }
                                        >
                                          <Form.Item
                                            label={'Key'}
                                            name={[name, 'key']}
                                          >
                                            <Input
                                              placeholder={intl.formatMessage({
                                                id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                                                defaultMessage: '请输入',
                                              })}
                                            />
                                          </Form.Item>
                                        </Col>
                                        <Col
                                          span={
                                            topologyConfiguration ===
                                            'nodeSelector'
                                              ? 8
                                              : 6
                                          }
                                        >
                                          <Form.Item
                                            label={'Operator'}
                                            name={[name, 'operator']}
                                          >
                                            <Select
                                              options={[
                                                ...(topologyConfiguration !==
                                                'Pod Affinity'
                                                  ? [
                                                      {
                                                        value: 'Equal',
                                                        label: 'Equal',
                                                      },
                                                    ]
                                                  : []),
                                                ...(topologyConfiguration !==
                                                'Toleration'
                                                  ? [
                                                      {
                                                        value: 'In',
                                                        label: 'In',
                                                      },
                                                      {
                                                        value: 'NotIn',
                                                        label: 'NotIn',
                                                      },
                                                      {
                                                        value: 'Exist',
                                                        label: 'Exist',
                                                      },
                                                    ]
                                                  : []),
                                                {
                                                  value: 'DoesNoExist',
                                                  label: 'DoesNoExist',
                                                },
                                              ]}
                                            />
                                          </Form.Item>
                                        </Col>
                                        <Col
                                          span={
                                            topologyConfiguration ===
                                            'nodeSelector'
                                              ? 8
                                              : 6
                                          }
                                        >
                                          <Form.Item
                                            label={'Value'}
                                            name={[name, 'value']}
                                          >
                                            {topologyConfiguration ===
                                            'Toleration' ? (
                                              <Input
                                                placeholder={intl.formatMessage(
                                                  {
                                                    id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                                                    defaultMessage: '请输入',
                                                  },
                                                )}
                                              />
                                            ) : (
                                              <Select maxTagCount={5} />
                                            )}
                                          </Form.Item>
                                        </Col>

                                        {topologyConfiguration ===
                                          'Toleration' && (
                                          <Col span={2}>
                                            <Form.Item
                                              label={'Seconds'}
                                              name={[name, 'seconds']}
                                            >
                                              <InputNumber
                                                min={1}
                                                placeholder={intl.formatMessage(
                                                  {
                                                    id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                                                    defaultMessage: '请输入',
                                                  },
                                                )}
                                              />
                                            </Form.Item>
                                          </Col>
                                        )}

                                        <DeleteOutlined
                                          style={{
                                            color: 'rgba(0, 0, 0, .45)',
                                            paddingRight: 8,
                                          }}
                                          onClick={() => {
                                            subOpt.remove(subField.name);
                                          }}
                                        />
                                      </Row>
                                    </div>
                                  ))}

                                  <Dropdown
                                    overlay={
                                      <Menu
                                        onClick={({ key }) => {
                                          setTopologyConfiguration(key);
                                          subOpt.add();
                                        }}
                                      >
                                        <Menu.Item key="nodeSelector">
                                          <span>nodeSelector</span>
                                        </Menu.Item>
                                        <Menu.Item key="Pod Affinity">
                                          <span>Pod Affinity</span>
                                        </Menu.Item>
                                        <Menu.Item key="Toleration">
                                          <span>Toleration</span>
                                        </Menu.Item>
                                      </Menu>
                                    }
                                  >
                                    <Button
                                      type="dashed"
                                      block
                                      icon={<PlusOutlined />}
                                    >
                                      Click to add topology configuration
                                      <DownOutlined />
                                    </Button>
                                  </Dropdown>
                                </div>
                              )}
                            </Form.List>
                          </Form.Item>
                        </>
                      )}
                    </div>
                  </Card>
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
