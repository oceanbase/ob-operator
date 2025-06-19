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

import { EFFECT_LIST, OPERATOR_LIST } from '@/constants/node';
import { RULER_ZONE } from '@/constants/rules';
import { getK8sObclusterListReq, getNodeLabelsReq } from '@/services';
import { useAccess } from '@umijs/max';
import { useRequest } from 'ahooks';
import { useEffect, useState } from 'react';

export default function Topo({ form }) {
  const { data: K8sClustersList } = useRequest(getK8sObclusterListReq);

  const [showTopology, setShowTopology] = useState<boolean>(false);
  const [formsubIndex, setFromSubIndex] = useState({});

  const [keyList, setKeyList] = useState([]);
  const [valList, setValList] = useState([]);

  const access = useAccess();

  const affinitiesOperatorList = [
    {
      value: 'In',
      label: 'In',
    },
    {
      value: 'NotIn',
      label: 'NotIn',
    },
    {
      value: 'Exists',
      label: 'Exists',
    },
    {
      value: 'DoesNoExist',
      label: 'DoesNoExist',
    },
  ];

  const basicFrom = (topologyConfiguration, name) => (
    <>
      <Col
        span={topologyConfiguration === 'NodeSelector' ? 7 : 5}
        style={{ paddingBottom: 24 }}
      >
        <Form.Item
          label={'Key'}
          name={[name, 'key']}
          rules={[
            {
              required: true,
              message: intl.formatMessage(
                {
                  id: 'src.pages.Cluster.New.AE7F8E70',
                  defaultMessage: '请选择 ${topologyConfiguration} Key',
                },
                { topologyConfiguration: topologyConfiguration },
              ),
            },
          ]}
        >
          {topologyConfiguration === 'NodeSelector' ? (
            <Select
              showSearch
              placeholder={intl.formatMessage({
                id: 'OBDashboard.components.NodeSelector.PleaseSelect',
                defaultMessage: '请选择',
              })}
              optionFilterProp="label"
              //@ts-expect-error Custom option component type is incompatible
              filterOption={filterOption}
              options={keyList}
              allowClear
            />
          ) : (
            <Input
              placeholder={intl.formatMessage({
                id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                defaultMessage: '请输入',
              })}
            />
          )}
        </Form.Item>
      </Col>
      <Col span={topologyConfiguration === 'NodeSelector' ? 8 : 6}>
        <Form.Item
          label={'Operator'}
          name={[name, 'operator']}
          rules={[
            {
              required: true,
              message: intl.formatMessage(
                {
                  id: 'src.pages.Cluster.New.5344AD47',
                  defaultMessage: '请选择 ${topologyConfiguration} Operator',
                },
                { topologyConfiguration: topologyConfiguration },
              ),
            },
          ]}
        >
          <Select
            placeholder={intl.formatMessage({
              id: 'OBDashboard.components.NodeSelector.PleaseSelect',
              defaultMessage: '请选择',
            })}
            options={
              topologyConfiguration === 'Tolerations'
                ? OPERATOR_LIST
                : affinitiesOperatorList
            }
          />
        </Form.Item>
      </Col>
      <Col span={topologyConfiguration === 'NodeSelector' ? 8 : 6}>
        <Form.Item
          label={topologyConfiguration === 'Tolerations' ? 'Value' : 'Values'}
          name={
            topologyConfiguration === 'Tolerations'
              ? [name, 'value']
              : [name, 'values']
          }
          rules={[
            {
              required: true,
              message: intl.formatMessage(
                {
                  id: 'src.pages.Cluster.New.870724D5',
                  defaultMessage: '请输入 ${topologyConfiguration} Value',
                },
                { topologyConfiguration: topologyConfiguration },
              ),
            },
          ]}
        >
          {topologyConfiguration === 'NodeSelector' ? (
            <Select
              showSearch
              mode="multiple"
              placeholder={intl.formatMessage({
                id: 'OBDashboard.components.NodeSelector.PleaseSelect',
                defaultMessage: '请选择',
              })}
              optionFilterProp="label"
              //@ts-expect-error Custom option component type is incompatible
              filterOption={filterOption}
              options={valList}
              allowClear
            />
          ) : topologyConfiguration === 'PodAffinity' ? (
            <Select
              mode="tags"
              style={{ width: '100%' }}
              tokenSeparators={[',']}
              options={[]}
              placeholder={intl.formatMessage({
                id: 'src.pages.Cluster.New.DB6FD585',
                defaultMessage: '输入后按回车添加',
              })}
              // 下拉框不展示
              dropdownStyle={{
                display: 'none',
                height: 0,
              }}
              // 下拉 icon 不展示
              suffixIcon={<></>}
            />
          ) : (
            <Input
              placeholder={intl.formatMessage({
                id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                defaultMessage: '请输入',
              })}
            />
          )}
        </Form.Item>
      </Col>
    </>
  );

  const filterOption = (
    input: string,
    option: { label: string; value: string },
  ) => (option?.label ?? '').toLowerCase().includes(input.toLowerCase());

  useEffect(() => {
    if (access.systemread || access.systemwrite) {
      const promise = getNodeLabelsReq();
      promise.then((data) => {
        setKeyList(data.key);
        setValList(data.value?.filter((item) => item.label !== ''));
      });
    }
  }, []);

  const NodeSelectorFrom: React.FC = ({ fieldName }) => (
    <Form.Item label="NodeSelector">
      <Form.List name={[fieldName, 'nodeSelector']}>
        {(subFields, { add, remove }) => (
          <>
            {subFields.map(({ key, name }) => {
              return (
                <div key={key}>
                  <Row gutter={8}>
                    {basicFrom('NodeSelector', name)}
                    <Form.Item
                      noStyle
                      name={[name, 'type']}
                      initialValue={'NODE'}
                    />

                    <DeleteOutlined
                      onClick={() => remove(name)}
                      style={{ marginBottom: 15 }}
                    />
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
                Add NodeSelector
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
                  <Col span={6}>
                    <Form.Item
                      label={'Type'}
                      name={[name, 'type']}
                      rules={[
                        {
                          required: true,
                          message: intl.formatMessage({
                            id: 'src.pages.Cluster.New.AC56EBD8',
                            defaultMessage: '请选择 PodAffinity Type',
                          }),
                        },
                      ]}
                    >
                      <Select
                        placeholder={intl.formatMessage({
                          id: 'OBDashboard.components.NodeSelector.PleaseSelect',
                          defaultMessage: '请选择',
                        })}
                        options={[
                          {
                            value: 'POD',
                            label: 'Affinity',
                          },
                          {
                            value: 'POD_ANTI',
                            label: 'AntiAffinity',
                          },
                        ]}
                      />
                    </Form.Item>
                  </Col>
                  {basicFrom('PodAffinity', name)}
                  <DeleteOutlined
                    onClick={() => remove(name)}
                    style={{ marginBottom: 15 }}
                  />
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
                  {basicFrom('Tolerations', name)}
                  <Col span={4}>
                    <Form.Item
                      label={'Effect'}
                      name={[name, 'effect']}
                      rules={[
                        {
                          required: true,
                          message: intl.formatMessage({
                            id: 'src.pages.Cluster.New.9DDB4176',
                            defaultMessage: '请选择 Tolerations Effect',
                          }),
                        },
                      ]}
                    >
                      <Select
                        placeholder={intl.formatMessage({
                          id: 'OBDashboard.components.NodeSelector.PleaseSelect',
                          defaultMessage: '请选择',
                        })}
                        options={EFFECT_LIST}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={2}>
                    <Form.Item
                      label={'Seconds'}
                      name={[name, 'tolerationSeconds']}
                    >
                      <InputNumber
                        min={1}
                        placeholder={intl.formatMessage({
                          id: 'OBDashboard.Cluster.New.Topo.PleaseEnter',
                          defaultMessage: '请输入',
                        })}
                      />
                    </Form.Item>
                  </Col>
                  <DeleteOutlined
                    onClick={() => remove(name)}
                    style={{ marginBottom: 15 }}
                  />
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

  const options = (K8sClustersList?.data || [])?.map((item) => ({
    value: item.name,
    label: item.name,
  }));

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
                    <Col span={3}>
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
                          style={{ width: '100%' }}
                          min={1}
                        />
                      </Form.Item>
                    </Col>
                    <Col span={5}>
                      <Form.Item
                        label={intl.formatMessage({
                          id: 'src.pages.Cluster.New.6CB28C7E',
                          defaultMessage: 'K8s 集群',
                        })}
                        name={[field.name, 'k8sCluster']}
                      >
                        <Select
                          showSearch
                          placeholder={intl.formatMessage({
                            id: 'src.pages.Cluster.New.9B629F24',
                            defaultMessage: '请选择 K8s 集群',
                          })}
                          optionFilterProp="label"
                          options={options}
                        />
                      </Form.Item>
                    </Col>
                    <Col span={3}>
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
                                    {!!nodeSelector &&
                                      `- ${nodeSelector?.length} node selector`}
                                  </div>
                                  <div>
                                    {!!affinities &&
                                      `- ${affinities?.length} pod affinity`}
                                  </div>
                                  <div>
                                    {!!tolerations &&
                                      `- ${tolerations?.length} tolerations`}
                                  </div>
                                </>
                              ) : (
                                '(empty)'
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
                        [field.key]: !showTopology,
                      });
                    }}
                  >
                    {formsubIndex[field.key] ? (
                      <DownOutlined />
                    ) : (
                      <RightOutlined />
                    )}
                    Topology
                  </Space>
                  {formsubIndex[field.key] && (
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
