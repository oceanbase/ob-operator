import { cluster, K8sClusterApi } from '@/api';
import { EFFECT_LIST, OPERATOR_LIST } from '@/constants/node';
import { intl } from '@/utils/intl';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { useRequest } from 'ahooks';
import {
  Button,
  Col,
  Drawer,
  Form,
  Input,
  message,
  Row,
  Select,
  Space,
  Tabs,
  TabsProps,
} from 'antd';
import { flattenDeep, uniqBy } from 'lodash';
import React, { useEffect, useState } from 'react';

export interface BatchEditNodeDrawerProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  initialValues: any[];
  name: string;
  namespace: string;
  selectedRowKeys: string[];
  k8sClusterName: string;
}

const BatchEditNodeDrawer: React.FC<BatchEditNodeDrawerProps> = ({
  visible,
  onCancel,
  onSuccess,
  selectedRowKeys,
  k8sClusterName,
}) => {
  const [form] = Form.useForm<API.CreateClusterData>();
  const { validateFields, resetFields } = form;
  const [tabKey, setTabKey] = useState<string>('labels');

  const onChange = (key: string) => {
    setTabKey(key);
  };

  const labelsOption = uniqBy(
    flattenDeep(selectedRowKeys?.map((item) => item.labels)),
    'key',
  ).map((item) => ({
    label: item.key,
    value: item.key,
  }));

  const taintsOption = uniqBy(
    flattenDeep(selectedRowKeys?.map((item) => item.taints)),
    'key',
  )?.map((item) => ({
    label: item.key,
    value: item.key,
  }));

  const { runAsync: batchUpdateK8sNodes, loading } = useRequest(
    cluster.batchUpdateK8sNodes,
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
          resetFields();
        }
      },
    },
  );

  const {
    runAsync: batchUpdateRemoteK8sNode,
    loading: batchUpdateK8sNodeLoading,
  } = useRequest(K8sClusterApi.batchUpdateRemoteK8sNode, {
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
        resetFields();
      }
    },
  });

  useEffect(() => {
    setTabKey('labels');
  }, [visible]);
  const basicForm = (title) => {
    const fromName = title === 'labels';
    return (
      <Form.List name={title}>
        {(fields, { add, remove }) => (
          <>
            {fields.map(({ key, name, ...restField }) => {
              return (
                <Row gutter={8}>
                  <Col span={fromName ? 6 : 4}>
                    <Form.Item
                      {...restField}
                      name={[name, 'operation']}
                      initialValue={'overwrite'}
                    >
                      <Select
                        options={[
                          {
                            value: 'overwrite',
                            label: intl.formatMessage({
                              id: 'src.pages.Overview.D7DF0F23',
                              defaultMessage: '新增/更新',
                            }),
                          },
                          {
                            value: 'delete',
                            label: intl.formatMessage({
                              id: 'src.pages.Overview.B3589C81',
                              defaultMessage: '删除',
                            }),
                          },
                        ]}
                      />
                    </Form.Item>
                  </Col>
                  <Col span={fromName ? 8 : 4}>
                    <Form.Item
                      noStyle
                      dependencies={[name, 'operation']}
                      shouldUpdate
                    >
                      {({ getFieldValue }) => {
                        return (
                          <Form.Item
                            {...restField}
                            name={[name, 'key']}
                            rules={[
                              {
                                required: true,
                                message: intl.formatMessage({
                                  id: 'src.pages.Overview.8666CAD2',
                                  defaultMessage: '请输入 Keys',
                                }),
                              },
                            ]}
                          >
                            {getFieldValue(title)[key]?.operation ===
                            'delete' ? (
                              <Select
                                showSearch
                                placeholder={intl.formatMessage({
                                  id: 'src.pages.Overview.71AA631F',
                                  defaultMessage: '请输入 Keys',
                                })}
                                optionFilterProp="label"
                                options={
                                  tabKey === 'labels'
                                    ? labelsOption
                                    : taintsOption
                                }
                              />
                            ) : (
                              <Input
                                placeholder={intl.formatMessage({
                                  id: 'src.pages.Overview.8A398AD4',
                                  defaultMessage: '请输入 Keys',
                                })}
                              />
                            )}
                          </Form.Item>
                        );
                      }}
                    </Form.Item>
                  </Col>
                  <Col span={fromName ? 10 : 16}>
                    <Form.Item
                      noStyle
                      dependencies={[name, 'operation']}
                      shouldUpdate
                    >
                      {({ getFieldValue }) => {
                        return (
                          <Row gutter={8}>
                            {getFieldValue(title)[key]?.operation !==
                            'delete' ? (
                              <>
                                {fromName && (
                                  <Col span={2}>
                                    <Form.Item {...restField}>=</Form.Item>
                                  </Col>
                                )}

                                {!fromName && (
                                  <Col span={5}>
                                    <Form.Item
                                      {...restField}
                                      name={[name, 'operator']}
                                      dependencies={[name, 'value']}
                                      initialValue={'Equal'}
                                    >
                                      <Select
                                        placeholder={intl.formatMessage({
                                          id: 'src.pages.Overview.AC07C9B6',
                                          defaultMessage: '请选择',
                                        })}
                                        options={OPERATOR_LIST}
                                      />
                                    </Form.Item>
                                  </Col>
                                )}
                                <Form.Item
                                  noStyle
                                  dependencies={[name, 'operator']}
                                  shouldUpdate
                                >
                                  {({ getFieldValue }) => {
                                    return (
                                      <Col span={fromName ? 18 : 8}>
                                        {(getFieldValue(title)[key]
                                          ?.operator === 'Equal' ||
                                          fromName) && (
                                          <Form.Item
                                            {...restField}
                                            name={[name, 'value']}
                                          >
                                            <Input
                                              placeholder={intl.formatMessage({
                                                id: 'src.pages.Overview.C35F8F39',
                                                defaultMessage: '请输入',
                                              })}
                                            />
                                          </Form.Item>
                                        )}
                                      </Col>
                                    );
                                  }}
                                </Form.Item>
                                {!fromName && (
                                  <Col span={1}>
                                    <Form.Item {...restField}>:</Form.Item>
                                  </Col>
                                )}
                                {!fromName && (
                                  <Col span={6}>
                                    <Form.Item
                                      {...restField}
                                      name={[name, 'effect']}
                                    >
                                      <Select options={EFFECT_LIST} />
                                    </Form.Item>
                                  </Col>
                                )}
                              </>
                            ) : (
                              <Col span={20}></Col>
                            )}
                            <Col span={1}>
                              <Form.Item {...restField}>
                                <DeleteOutlined onClick={() => remove(name)} />
                              </Form.Item>
                            </Col>
                          </Row>
                        );
                      }}
                    </Form.Item>
                  </Col>
                </Row>
              );
            })}
            <Form.Item>
              <Button
                type="dashed"
                onClick={() => add()}
                block
                icon={<PlusOutlined />}
              >
                {intl.formatMessage({
                  id: 'src.pages.Overview.4ADD12B7',
                  defaultMessage: '添加',
                })}
              </Button>
            </Form.Item>
          </>
        )}
      </Form.List>
    );
  };

  const items: TabsProps['items'] = [
    {
      key: 'labels',
      label: 'Labels',
    },
    {
      key: 'taints',
      label: 'Taints',
    },
  ];

  return (
    <Drawer
      title={intl.formatMessage({
        id: 'src.pages.Overview.684169FA',
        defaultMessage: '批量编辑节点',
      })}
      open={visible}
      destroyOnClose
      onClose={() => {
        onCancel();
        resetFields();
      }}
      width={800}
      footer={
        <Space>
          <Button
            onClick={() => {
              onCancel();
              resetFields();
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.3B8C3AE9',
              defaultMessage: '取消',
            })}
          </Button>
          <Button
            type="primary"
            loading={k8sClusterName ? batchUpdateK8sNodeLoading : loading}
            onClick={() => {
              validateFields().then((values) => {
                const { labels, taints } = values;
                const labelOperations = labels;
                const taintOperations = taints?.map((item) => ({
                  key: item.key,
                  value: item.operator === 'Equal' ? item.value : undefined,
                  effect: item.effect,
                  operation: item.operation,
                }));

                const nodes = selectedRowKeys.map((item) => item.name);
                const final = { nodes, labelOperations, taintOperations };

                if (k8sClusterName) {
                  batchUpdateRemoteK8sNode(k8sClusterName, final);
                } else {
                  batchUpdateK8sNodes(final);
                }
              });
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.AC4C9FB4',
              defaultMessage: '确定',
            })}
          </Button>
        </Space>
      }
    >
      <Tabs items={items} onChange={onChange} />
      <Form
        form={form}
        layout="vertical"
        initialValues={{
          labels: [{ operation: 'overwrite' }],
          taints: [{ operation: 'overwrite', operator: 'Equal' }],
        }}
      >
        {basicForm(tabKey)}
      </Form>
    </Drawer>
  );
};

export default BatchEditNodeDrawer;
