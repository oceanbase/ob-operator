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
  Form,
  Input,
  InputNumber,
  Popconfirm,
  Row,
  Select,
  Space,
} from 'antd';

import { RULER_ZONE } from '@/constants/rules';
import { useState } from 'react';

export default function Topo({ form }) {
  const [showTopology, setShowTopology] = useState<boolean>(false);
  const [formsubIndex, setFromSubIndex] = useState({});

  const basicFrom = (topologyConfiguration, name) => (
    <>
      <Col
        span={topologyConfiguration === 'nodeSelector' ? 7 : 5}
        style={{ paddingBottom: 24 }}
      >
        <Form.Item label={'Key'} name={[name, 'key']}>
          <Input
            placeholder={intl.formatMessage({
              id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
              defaultMessage: '请输入',
            })}
          />
        </Form.Item>
      </Col>
      <Col span={topologyConfiguration === 'nodeSelector' ? 8 : 6}>
        <Form.Item label={'Operator'} name={[name, 'operator']}>
          <Select
            options={[
              ...(topologyConfiguration !== 'Pod Affinity'
                ? [
                    {
                      value: 'Equal',
                      label: 'Equal',
                    },
                  ]
                : []),
              ...(topologyConfiguration !== 'Toleration'
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
      <Col span={topologyConfiguration === 'nodeSelector' ? 8 : 6}>
        <Form.Item label={'Value'} name={[name, 'value']}>
          {topologyConfiguration === 'Toleration' ? (
            <Input
              placeholder={intl.formatMessage({
                id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                defaultMessage: '请输入',
              })}
            />
          ) : (
            <Select maxTagCount={5} />
          )}
        </Form.Item>
      </Col>
    </>
  );

  const NodeSelectorFrom: React.FC = ({ fieldName }) => (
    <Form.Item label="nodeSelector">
      <Form.List name={[fieldName, 'nodeSelector']}>
        {(subFields, { add, remove }) => (
          <>
            {subFields.map(({ key, name }) => {
              return (
                <div key={key}>
                  <Row gutter={8}>
                    {basicFrom('nodeSelector', name)}
                    <DeleteOutlined onClick={() => remove(name)} />
                  </Row>
                </div>
              );
            })}
            <Form.Item noStyle>
              <Button
                style={{ marginBottom: 24 }}
                type="dashed"
                onClick={() => add()}
                block
                icon={<PlusOutlined />}
              >
                Add nodeSelector
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>
    </Form.Item>
  );

  const PodAffinityFrom: React.FC = ({ fieldName }) => (
    <Form.Item label="PodAffinity">
      <Form.List name={[fieldName, 'affinities']}>
        {(Podfields, { add, remove }) => (
          <>
            {Podfields.map(({ key, name }) => (
              <div key={key}>
                <Row gutter={8}>
                  {basicFrom('PodAffinity', name)}
                  <Col span={6}>
                    <Form.Item label={'Type'} name={[name, 'type']}>
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

                  <DeleteOutlined onClick={() => remove(name)} />
                </Row>
              </div>
            ))}
            <Form.Item noStyle>
              <Button
                style={{ marginBottom: 24 }}
                type="dashed"
                onClick={() => add()}
                block
                icon={<PlusOutlined />}
              >
                Add PodAffinity
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>
    </Form.Item>
  );

  const TolerationFrom: React.FC = ({ fieldName }) => (
    <Form.Item label="Tolerations">
      <Form.List name={[fieldName, 'tolerations']}>
        {(tolerationfields, { add, remove }) => (
          <>
            {tolerationfields.map(({ key, name }) => (
              <div key={key}>
                <Row gutter={8}>
                  {basicFrom('Toleration', name)}
                  <Col span={4}>
                    <Form.Item label={'Effect'} name={[name, 'Effect']}>
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
                  <Col span={2}>
                    <Form.Item label={'Seconds'} name={[name, 'seconds']}>
                      <InputNumber
                        min={1}
                        placeholder={intl.formatMessage({
                          id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                          defaultMessage: '请输入',
                        })}
                      />
                    </Form.Item>
                  </Col>
                  <DeleteOutlined onClick={() => remove(name)} />
                </Row>
              </div>
            ))}
            <Form.Item>
              <Button
                type="dashed"
                onClick={() => add()}
                block
                icon={<PlusOutlined />}
              >
                Add Toleration
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>
    </Form.Item>
  );

  return (
    <Col span={24}>
      <Card
        title={intl.formatMessage({
          id: 'dashboard.Cluster.New.Topo.Topology',
          defaultMessage: '拓扑',
        })}
      >
        <Form.List name="topology">
          {(fields, { add, remove }) => (
            <div>
              {fields.map((field) => (
                <Card style={{ marginBottom: 16 }} key={field.key}>
                  <Row gutter={8}>
                    <Col span={12}>
                      <Form.Item
                        label={intl.formatMessage({
                          id: 'OBDashboard.Cluster.New.Topo.ZoneName',
                          defaultMessage: 'Zone名称',
                        })}
                        validateFirst
                        name={[field.name, 'zone']}
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
                    <Col span={6}>
                      <Form.Item
                        label={intl.formatMessage({
                          id: 'OBDashboard.Cluster.New.Topo.NumberOfServers',
                          defaultMessage: 'server数',
                        })}
                        name={[field.name, 'replicas']}
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
                      <Form.Item label={'Topology summary'} shouldUpdate>
                        {() => {
                          const { topology } = form.getFieldsValue();

                          const content =
                            topology?.length > 0 ? topology[field.name] : null;

                          const { nodeSelector, affinities, tolerations } =
                            content || {};

                          return (
                            <div>
                              {!!nodeSelector ||
                              !!affinities ||
                              !!tolerations ? (
                                <>
                                  <div>
                                    {!!nodeSelector && '- nodeSelector'}
                                  </div>
                                  <div>{!!affinities && '- affinities'}</div>
                                  <div>{!!tolerations && '- tolerations'}</div>
                                </>
                              ) : (
                                '-'
                              )}
                            </div>
                          );
                        }}
                      </Form.Item>
                    </Col>
                    <Col span={1}>
                      {fields.length > 1 && (
                        <Form.Item
                          label={<></>}
                          style={{ marginBottom: 0 }}
                          name={[field.name, ' ']}
                        >
                          <Popconfirm
                            placement="left"
                            title={intl.formatMessage({
                              id: 'OBDashboard.Cluster.New.Topo.AreYouSureYouWant',
                              defaultMessage: '确定要删除该配置项吗？',
                            })}
                            onConfirm={() => {
                              remove(field.name);
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
                  <Space
                    onClick={() => {
                      setShowTopology(!showTopology);
                      setFromSubIndex({
                        ...formsubIndex,
                        [field.index]: !showTopology,
                      });
                    }}
                  >
                    {formsubIndex[field.index] ? (
                      <DownOutlined />
                    ) : (
                      <RightOutlined />
                    )}
                    Topology
                  </Space>
                  {formsubIndex[field.index] && (
                    <div style={{ marginTop: 16 }}>
                      <NodeSelectorFrom fieldName={field.name} />
                      <PodAffinityFrom fieldName={field.name} />
                      <TolerationFrom fieldName={field.name} />
                    </div>
                  )}
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
            </div>
          )}
        </Form.List>
      </Card>
    </Col>
  );
}
