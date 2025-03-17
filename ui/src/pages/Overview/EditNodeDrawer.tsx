import { cluster } from '@/api';
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
  Row,
  Select,
  Space,
  Tabs,
  TabsProps,
  message,
} from 'antd';
import React, { useState } from 'react';

export interface ParametersModalProps {
  visible: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  initialValues: any[];
  name: string;
  namespace: string;
}

const EditNodeDrawer: React.FC<ParametersModalProps> = ({
  visible,
  onCancel,
  nodeRecord,
  onSuccess,
  labels,
  taints,
}) => {
  const [form] = Form.useForm<API.CreateClusterData>();
  const { validateFields, resetFields } = form;
  const update = useUpdate();
  const [tabKey, setTabKey] = useState<string>('labels');

  const taintsIn = taints?.map((item) => ({
    key: item.key,
    value: item.value,
    operator: item.value ? 'Equal' : 'Exists',
    effect: item.effect,
  }));

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
        }
      },
    },
  );

  const onChange = (key: string) => {
    setTabKey(key);
  };

  const basicForm = (title) => {
    const fromName = title === 'labels';
    return (
      <Form.List name={title}>
        {(fields, { add, remove }) => (
          <>
            {fields.map(({ key, name, ...restField }) => {
              return (
                <Row gutter={16}>
                  <Col span={fromName ? 10 : 6}>
                    <Form.Item
                      {...restField}
                      name={[name, 'key']}
                      rules={[{ required: true, message: '请输入 Keys' }]}
                      label={key === 0 && 'Key'}
                    >
                      <Input placeholder="First Name" />
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
                            >
                              <Select
                                placeholder={'请选择'}
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
                    {({ getFieldValue }) => {
                      return (
                        <Col span={fromName ? 10 : 5}>
                          {getFieldValue(title)[key]?.operator !== 'Exists' && (
                            <Form.Item
                              {...restField}
                              name={[name, 'value']}
                              label={key === 0 && 'Value'}
                              rules={[
                                ...((fromName &&
                                  getFieldValue(title)[key]?.value) ||
                                getFieldValue(title)[key]?.operator === 'Equal'
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
      label: 'labels',
    },
    {
      key: 'taints',
      label: 'taints',
    },
  ];

  return (
    <Drawer
      title={'编辑节点'}
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
            loading={loading || taintsLoading}
            onClick={() => {
              validateFields().then((values) => {
                const name = nodeRecord?.name;
                if (tabKey === 'labels') {
                  putK8sNodeLabels(name, values);
                } else {
                  const obj = values.taints?.map((item) => ({
                    key: item.key,
                    value: item.operator === 'Equal' ? item.value : undefined,
                    effect: item.effect,
                  }));
                  putK8sNodeTaints(name, { taints: obj });
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
        initialValues={{ labels: labels, taints: taintsIn }}
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
