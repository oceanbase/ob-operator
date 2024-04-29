import { useModel } from '@umijs/max';
import { Card, Col, DatePicker, Form, Input, Row, Select, Tag } from 'antd';
import type { DefaultOptionType, SelectProps } from 'antd/es/select';
import { flatten } from 'lodash';

const LEVER_OPTIONS: SelectProps['options'] = [
  {
    label: '停服',
    value: '1',
  },
  {
    label: '严重',
    value: '2',
  },
  {
    label: '警告',
    value: '3',
  },
  {
    label: '注意',
    value: '4',
  },
  {
    label: '提醒',
    value: '5',
  },
];
const COLOR_MAP = {
  1: 'purple',
  2: 'red',
  3: 'gold',
  4: 'blue',
  5: 'green',
};
export default function Event() {
  const { clusterList, tenantList } = useModel('alarm');
  console.log('clusterList', clusterList);

  const OBJECT_OPTIONS: DefaultOptionType[] = [
    {
      label: '集群',
      value: 'cluster',
    },
    {
      label: '租户',
      value: 'tenant',
    },
    {
      label: 'OBServer',
      value: 'observer',
    },
  ];
  const onOk = (
    value: DatePickerProps['value'] | RangePickerProps['value'],
  ) => {
    console.log('onOk: ', value);
  };

  const getOptionsFromType = (
    type: 'cluster' | 'tenant' | 'observer' | undefined,
  ) => {
    if (type === 'cluster') {
      return clusterList?.map((cluster) => ({
        value: cluster.clusterName,
        label: cluster.clusterName,
      }));
    }
    if (type === 'tenant') {
      return clusterList?.map((cluster) => ({
        label: <span>{cluster.clusterName}</span>,
        title: cluster.clusterName,
        options: tenantList
          ?.filter(
            (tenant) =>
              tenant.namespace === cluster.namespace &&
              tenant.clusterResourceName === cluster.name,
          )
          .map((tenant) => ({
            label: tenant.tenantName,
            value: tenant.tenantName,
          })),
      }));
    }
    if (type === 'observer') {
      return clusterList?.map((cluster) => ({
        label: <span>{cluster.clusterName}</span>,
        title: cluster.clusterName,
        options: flatten(
          cluster.topology.map((zone) =>
            zone.observers.map((server) => ({
              label: server.address,
              value: server.address,
            })),
          ),
        ),
      }));
    }
    return [];
  };

  return (
    <div>
      <Card>
        <Form>
          <Row>
            <Col span={8}>
              <Form.Item
                wrapperCol={{ span: 18 }}
                labelCol={{ span: 9 }}
                label="对象类型"
                name={'objectType'}
              >
                <Select
                  allowClear
                  placeholder="请选择"
                  options={OBJECT_OPTIONS}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item noStyle dependencies={['objectType']}>
                {({ getFieldValue }) => {
                  return (
                    <Form.Item
                      wrapperCol={{ span: 18 }}
                      labelCol={{ span: 9 }}
                      label="告警对象"
                    >
                      <Select
                        style={{ width: '100%' }}
                        options={getOptionsFromType(
                          getFieldValue(['objectType']),
                        )}
                        placeholder="请选择"
                      />
                    </Form.Item>
                  );
                }}
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                wrapperCol={{ span: 18 }}
                labelCol={{ span: 9 }}
                label="告警等级"
                name={'level'}
              >
                <Select
                  mode="multiple"
                  allowClear
                  placeholder="请选择"
                  options={LEVER_OPTIONS?.map((item) => ({
                    value: item.value,
                    label: (
                      <Tag color={COLOR_MAP[item.value]}>{item.label}</Tag>
                    ),
                  }))}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                wrapperCol={{ span: 18 }}
                labelCol={{ span: 9 }}
                label="关键词"
                name={'keyWord'}
              >
                <Input placeholder="请输入关键词" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                wrapperCol={{ span: 18 }}
                labelCol={{ span: 9 }}
                label="开始时间"
              >
                <DatePicker
                  style={{ width: '100%' }}
                  placeholder="请选择"
                  showTime
                  onChange={(value, dateString) => {
                    console.log('Selected Time: ', value);
                    console.log('Formatted Selected Time: ', dateString);
                  }}
                  onOk={onOk}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                wrapperCol={{ span: 18 }}
                labelCol={{ span: 9 }}
                label="结束时间"
              >
                <DatePicker
                  placeholder="请选择"
                  style={{ width: '100%' }}
                  showTime
                  onChange={(value, dateString) => {
                    console.log('Selected Time: ', value);
                    console.log('Formatted Selected Time: ', dateString);
                  }}
                  onOk={onOk}
                />
              </Form.Item>
            </Col>
          </Row>
        </Form>
      </Card>
    </div>
  );
}
