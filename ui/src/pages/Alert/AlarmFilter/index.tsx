import {
  SERVERITY_MAP,
  LEVER_OPTIONS_ALARM,
  OBJECT_OPTIONS_ALARM,
} from '@/constants';
import { DownOutlined, UpOutlined } from '@ant-design/icons';
import { useModel } from '@umijs/max';
import { useUpdateEffect } from 'ahooks';
import type { FormInstance } from 'antd';
import { Button, Col, DatePicker, Form, Input, Row, Select, Tag } from 'antd';
import { flatten } from 'lodash';
import { useEffect, useState } from 'react';

interface AlarmFilterProps {
  form: FormInstance<unknown>;
  type: 'event' | 'shield' | 'rules';
}

const DEFAULT_VISIBLE_CONFIG = {
  objectType: true,
  object: true,
  level: true,
  keyword: true,
  startTime: true,
  endTime: true,
};

export default function AlarmFilter({ form, type }: AlarmFilterProps) {
  const { clusterList, tenantList } = useModel('alarm');
  const [isExpand, setIsExpand] = useState(true);
  const [visibleConfig, setVisibleConfig] = useState(DEFAULT_VISIBLE_CONFIG);
  const getOptionsFromType = (
    type: 'obcluster' | 'obtenant' | 'observer' | 'obzone' | undefined,
  ) => {
    if (type === 'obcluster') {
      return clusterList?.map((cluster) => ({
        value: cluster.clusterName,
        label: cluster.clusterName,
      }));
    }
    if (type === 'obtenant') {
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

  useEffect(() => {
    if (type === 'event') {
      setVisibleConfig({
        objectType: true,
        object: true,
        level: true,
        keyword: true,
        startTime: true,
        endTime: true,
      });
    }
    if (type === 'shield') {
      setVisibleConfig({
        objectType: true,
        object: true,
        keyword: true,
        startTime: false,
        endTime: false,
        level: false,
      });
    }
    if (type === 'rules') {
      setVisibleConfig({
        objectType: true,
        level: true,
        keyword: true,
        startTime: false,
        endTime: false,
        object: false,
      });
    }
  }, [type]);

  useUpdateEffect(() => {
    if (isExpand) {
      setVisibleConfig({
        objectType: true,
        object: true,
        level: true,
        keyword: true,
        startTime: true,
        endTime: true,
      });
    } else {
      setVisibleConfig({
        objectType: true,
        object: true,
        level: true,
        keyword: false,
        startTime: false,
        endTime: false,
      });
    }
  }, [isExpand]);

  return (
    <Form form={form}>
      <Row>
        {visibleConfig.objectType && (
          <Col span={8}>
            <Form.Item
              wrapperCol={{ span: 18 }}
              labelCol={{ span: 9 }}
              label="对象类型"
              name={['instance', 'type']}
            >
              <Select
                allowClear
                placeholder="请选择"
                options={OBJECT_OPTIONS_ALARM}
              />
            </Form.Item>
          </Col>
        )}

        {visibleConfig.object && (
          <Col span={8}>
            <Form.Item noStyle dependencies={[['instance', 'type']]}>
              {({ getFieldValue }) => {
                return (
                  <Form.Item
                    wrapperCol={{ span: 18 }}
                    labelCol={{ span: 9 }}
                    name={['instance', getFieldValue(['instance', 'type'])]}
                    label={type === 'event' ? '告警对象' : '屏蔽对象'}
                  >
                    <Select
                      style={{ width: '100%' }}
                      options={getOptionsFromType(
                        getFieldValue(['instance', 'type']),
                      )}
                      placeholder="请选择"
                    />
                  </Form.Item>
                );
              }}
            </Form.Item>
          </Col>
        )}

        {visibleConfig.level && (
          <Col span={8}>
            <Form.Item
              wrapperCol={{ span: 18 }}
              labelCol={{ span: 9 }}
              label="告警等级"
              name={'serverity'}
            >
              <Select
                mode="multiple"
                allowClear
                placeholder="请选择"
                options={LEVER_OPTIONS_ALARM?.map((item) => ({
                  value: item.value,
                  label: (
                    <Tag color={SERVERITY_MAP[item.value]?.color}>{item.label}</Tag>
                  ),
                }))}
              />
            </Form.Item>
          </Col>
        )}

        {visibleConfig.keyword && (
          <Col span={8}>
            <Form.Item
              wrapperCol={{ span: 18 }}
              labelCol={{ span: 9 }}
              label="关键词"
              name={'keyword'}
            >
              <Input placeholder="请输入关键词" />
            </Form.Item>
          </Col>
        )}

        {visibleConfig.startTime && (
          <Col span={8}>
            <Form.Item
              wrapperCol={{ span: 18 }}
              labelCol={{ span: 9 }}
              label="开始时间"
              name={'startTime'}
            >
              <DatePicker
                style={{ width: '100%' }}
                placeholder="请选择"
                showTime
                onChange={(value, dateString) => {
                  console.log('Selected Time: ', value);
                  console.log('Formatted Selected Time: ', dateString);
                }}
              />
            </Form.Item>
          </Col>
        )}

        {visibleConfig.endTime && (
          <Col span={8}>
            <Form.Item
              name={'endTime'}
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
              />
            </Form.Item>
          </Col>
        )}
      </Row>
      <div style={{ float: 'right' }}>
        <Button type="link" onClick={() => form.resetFields()}>
          重置
        </Button>
        {type === 'event' &&
          (isExpand ? (
            <Button onClick={() => setIsExpand(false)} type="link">
              收起
              <UpOutlined />
            </Button>
          ) : (
            <Button onClick={() => setIsExpand(true)} type="link">
              展开
              <DownOutlined />
            </Button>
          ))}
      </div>
    </Form>
  );
}
