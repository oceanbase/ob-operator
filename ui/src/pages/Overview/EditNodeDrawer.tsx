import { cluster, K8sClusterApi } from '@/api';
import { EFFECT_LIST, OPERATOR_LIST } from '@/constants/node';
import { intl } from '@/utils/intl';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { useRequest, useUpdate } from 'ahooks';
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
import React, { useEffect, useState } from 'react';

export interface ParametersModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  initialValues: any[];
  name: string;
  namespace: string;
  k8sClusterName: string;
  type: string;
}

const EditNodeDrawer: React.FC<ParametersModalProps> = ({
  visible,
  onCancel,
  nodeRecord,
  onSuccess,
  k8sClusterName,
  type,
}) => {
  const [form] = Form.useForm<API.CreateClusterData>();
  const { validateFields, resetFields, setFieldsValue } = form;
  const update = useUpdate();
  const [tabKey, setTabKey] = useState<string>('labels');
  const { labels, taints } = nodeRecord;

  const formatTaints = (value) => {
    return value?.map((item) => ({
      key: item.key,
      value: item.value,
      operator: item.value ? 'Equal' : 'Exists',
      effect: item.effect,
    }));
  };

  useEffect(() => {
    if (visible) {
      setFieldsValue({
        labels: labels,
        taints: formatTaints(taints),
      });
    }
    setTabKey('labels');
  }, [visible]);

  const { runAsync: putK8sNodeLabels, loading } = useRequest(
    cluster.putK8sNodeLabels,
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
          setFieldsValue({
            labels: res?.data?.info?.labels,
          });
        }
      },
    },
  );

  const { runAsync: putK8sNodeTaints, loading: taintsLoading } = useRequest(
    cluster.putK8sNodeTaints,
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
          setFieldsValue({
            labels: res?.data?.info?.labels,
            taints: formatTaints(res?.data?.info?.taints),
          });
          setTabKey('labels');
        }
      },
    },
  );

  const {
    runAsync: putRemoteK8sNodeLabels,
    loading: putRemoteK8sNodeLabelsLoading,
  } = useRequest(K8sClusterApi.putRemoteK8sNodeLabels, {
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

  const {
    runAsync: putRemoteK8sNodeTaints,
    loading: putRemoteK8sNodeTaintsLoading,
  } = useRequest(K8sClusterApi.putRemoteK8sNodeTaints, {
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

  const onChange = (key: string) => {
    setTabKey(key);
  };

  const basicForm = (title) => {
    const fromName = title === 'labels';
    return (
      <Form.List name={title}>
        {(fields, { add, remove }) => (
          <>
            {fields.map(({ key, name, ...restField }, index) => {
              return (
                <Row gutter={16}>
                  <Col span={fromName ? 10 : 6}>
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[
                        {
                          required: true,
                          message: intl.formatMessage({
                            id: 'src.pages.Overview.ECC5714F',
                            defaultMessage: '请输入 Keys',
                          }),
                        },
                      ]}
                      label={key === 0 && 'Key'}
                    >
                      <Input
                        placeholder={intl.formatMessage({
                          id: 'src.pages.Overview.8B049D70',
                          defaultMessage: '请输入 ',
                        })}
                      />
                    </Form.Item>
                  </Col>
                  {fromName && (
                    <Col span={1}>
                      <Form.Item {...restField} label={key === 0 && <></>}>
                        =
                      </Form.Item>
                    </Col>
                  )}
                  {!fromName && (
                    <Col span={5}>
                      <Form.Item noStyle dependencies={[name, 'value']}>
                        {({ getFieldValue }) => {
                          return (
                            <Form.Item
                              {...restField}
                              label={key === 0 && <></>}
                              name={[name, 'operator']}
                              initialValue={'Equal'}
                            >
                              <Select
                                placeholder={intl.formatMessage({
                                  id: 'src.pages.Overview.0736D5A2',
                                  defaultMessage: '请选择',
                                })}
                                defaultValue={
                                  getFieldValue(title)[key]?.value
                                    ? 'Equal'
                                    : 'Exists'
                                }
                                options={OPERATOR_LIST}
                              />
                            </Form.Item>
                          );
                        }}
                      </Form.Item>
                    </Col>
                  )}
                  <Form.Item
                    noStyle
                    dependencies={[name, 'operator']}
                    shouldUpdate
                  >
                    {() => {
                      const currentValues = form.getFieldsValue();
                      return (
                        <Col span={fromName ? 10 : 5}>
                          {(fromName ||
                            currentValues?.[title]?.[index]?.operator ===
                              'Equal') && (
                            <Form.Item
                              {...restField}
                              name={[name, 'value']}
                              label={key === 0 && 'Value'}
                              rules={[
                                ...(currentValues?.[title]?.[index]
                                  ?.operator === 'Equal'
                                  ? [
                                      {
                                        required: true,
                                        message: intl.formatMessage({
                                          id: 'src.pages.Overview.8926036C',
                                          defaultMessage: '请输入 Value',
                                        }),
                                      },
                                    ]
                                  : []),
                              ]}
                            >
                              <Input
                                placeholder={intl.formatMessage({
                                  id: 'src.pages.Overview.F0C02D08',
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
                      <Form.Item {...restField} label={key === 0 && <></>}>
                        :
                      </Form.Item>
                    </Col>
                  )}
                  {!fromName && (
                    <Col span={6}>
                      <Form.Item
                        {...restField}
                        label={key === 0 && 'Effect'}
                        name={[name, 'effect']}
                      >
                        <Select options={EFFECT_LIST} />
                      </Form.Item>
                    </Col>
                  )}

                  <Col span={1}>
                    <Form.Item {...restField} label={key === 0 && <></>}>
                      <DeleteOutlined onClick={() => remove(name)} />
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
                  id: 'src.pages.Overview.833ECBA2',
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
        id: 'src.pages.Overview.406A1FCA',
        defaultMessage: '编辑节点',
      })}
      open={visible}
      destroyOnClose
      onClose={() => {
        onCancel();
        resetFields();
        setTabKey('labels');
      }}
      width={700}
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
            loading={
              loading ||
              taintsLoading ||
              putRemoteK8sNodeTaintsLoading ||
              putRemoteK8sNodeLabelsLoading
            }
            onClick={() => {
              validateFields().then((values) => {
                const name = nodeRecord?.name;
                if (tabKey === 'labels') {
                  if (type === 'k8s') {
                    putRemoteK8sNodeLabels(k8sClusterName, name, values);
                  } else {
                    putK8sNodeLabels(name, values);
                  }
                } else {
                  const obj = values.taints?.map((item) => ({
                    key: item.key,
                    value: item.operator === 'Equal' ? item.value : undefined,
                    effect: item.effect,
                  }));
                  if (type === 'k8s') {
                    putRemoteK8sNodeTaints(k8sClusterName, name, {
                      taints: obj,
                    });
                  } else {
                    putK8sNodeTaints(name, { taints: obj });
                  }
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
        onValuesChange={() => {
          update();
        }}
      >
        {basicForm(tabKey)}
      </Form>
    </Drawer>
  );
};

export default EditNodeDrawer;
