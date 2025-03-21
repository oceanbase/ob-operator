import { cluster } from '@/api';
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
  Row,
  Select,
  Space,
  Tabs,
  TabsProps,
  message,
} from 'antd';
import { flattenDeep, uniqBy } from 'lodash';
import React, { useState } from 'react';

export interface BatchEditNodeDrawerProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  initialValues: any[];
  name: string;
  namespace: string;
  selectedRowKeys: string[];
}

const BatchEditNodeDrawer: React.FC<BatchEditNodeDrawerProps> = ({
  visible,
  onCancel,
  onSuccess,
  selectedRowKeys,
}) => {
  const [form] = Form.useForm<API.CreateClusterData>();
  const { validateFields, resetFields } = form;
  const [tabKey, setTabKey] = useState<string>('labels');

  const onChange = (key: string) => {
    setTabKey(key);
  };

  const labelsOption = flattenDeep(
    uniqBy(
      selectedRowKeys?.map((item) => item.labels),
      'key',
    ),
  )?.map((item) => ({
    label: item.key,
    value: item.key,
  }));

  const taintsOption = flattenDeep(
    uniqBy(
      selectedRowKeys?.map((item) => item.taints),
      'key',
    ),
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
          setTabKey('labels');
        }
      },
    },
  );

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
                    <Form.Item {...restField} name={[name, 'operation']}>
                      <Select
                        options={[
                          {
                            value: 'overwrite',
                            label: '新增/更新',
                          },
                          {
                            value: 'delete',
                            label: '删除',
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
                                message: '请输入 Keys',
                              },
                            ]}
                          >
                            {getFieldValue(title)[key]?.operation ===
                            'delete' ? (
                              <Select
                                showSearch
                                placeholder="请输入 Keys"
                                optionFilterProp="label"
                                options={
                                  tabKey === 'labels'
                                    ? labelsOption
                                    : taintsOption
                                }
                              />
                            ) : (
                              <Input placeholder="请输入 Keys" />
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
                                    >
                                      <Select
                                        placeholder={'请选择'}
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
                                            rules={[
                                              ...((fromName &&
                                                getFieldValue(title)[key]
                                                  ?.value) ||
                                              getFieldValue(title)[key]
                                                ?.operator === 'Equal'
                                                ? [
                                                    {
                                                      required: true,
                                                      message: '请输入 Values',
                                                    },
                                                  ]
                                                : []),
                                            ]}
                                          >
                                            <Input placeholder="请输入" />
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
                添加
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
      title={'批量编辑节点'}
      open={visible}
      destroyOnClose
      onClose={() => {
        onCancel();
        resetFields();
        setTabKey('labels');
      }}
      width={800}
      footer={
        <Space>
          <Button
            onClick={() => {
              onCancel();
              resetFields();
              setTabKey('labels');
            }}
          >
            {intl.formatMessage({
              id: 'src.pages.Cluster.Detail.Overview.3B8C3AE9',
              defaultMessage: '取消',
            })}
          </Button>
          <Button
            type="primary"
            loading={loading}
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

                batchUpdateK8sNodes(final);
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
