import type { OceanbaseOBInstanceType } from '@/api/generated';
import {
  LEVER_OPTIONS_ALARM,
  OBJECT_OPTIONS_ALARM,
  SEVERITY_MAP,
} from '@/constants';
import { Alert } from '@/type/alert';
import { DownOutlined, UpOutlined } from '@ant-design/icons';
import { useModel } from '@umijs/max';
import { useDebounceFn, useUpdateEffect } from 'ahooks';
import type { FormInstance } from 'antd';
import { Button, Col, DatePicker, Form, Input, Row, Select, Tag } from 'antd';
import { useEffect, useState } from 'react';
import { getSelectList } from '../helper';

interface AlarmFilterProps {
  form: FormInstance<unknown>;
  depend: (body: unknown) => void;
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

export default function AlarmFilter({ form, type, depend }: AlarmFilterProps) {
  const { clusterList, tenantList } = useModel('alarm');
  const [isExpand, setIsExpand] = useState(true);
  const [visibleConfig, setVisibleConfig] = useState(DEFAULT_VISIBLE_CONFIG);
  const getOptionsFromType = (type: OceanbaseOBInstanceType) => {
    if (!type || !clusterList || (type === 'obtenant' && !tenantList))
      return [];
    const list = getSelectList(clusterList, type, tenantList);
    if (type === 'obcluster') {
      return list?.map((clusterName) => ({
        value: clusterName,
        label: clusterName,
      }));
    }
    if (type === 'obtenant') {
      return (list as Alert.TenantsList[])
        .map((cluster) => ({
          label: <span>{cluster.clusterName}</span>,
          title: cluster.clusterName,
          options: cluster.tenants?.map((item) => ({
            value: item,
            label: item,
          })),
        }))
        .filter((item) => item.options?.length);
    }
    if (type === 'observer') {
      return (list as Alert.ServersList[])
        .map((cluster) => ({
          label: <span>{cluster.clusterName}</span>,
          title: cluster.clusterName,
          options: cluster.servers?.map((item) => ({
            value: item,
            label: item,
          })),
        }))
        .filter((item) => item.options?.length);
    }
  };
  const { run: debounceDepend } = useDebounceFn(depend, { wait: 500 });
  const formData: any = Form.useWatch([], form);

  const findClusterName = (
    list: API.SimpleClusterList | API.TenantDetail[],
    type: 'obtenant' | 'observer',
    target: string,
  ) => {
    if (type === 'observer') {
      return (
        (list as API.SimpleClusterList).find((cluster) => {
          return cluster.topology.some((zone) =>
            zone.observers.some((server) => server.address === target),
          );
        })?.clusterName || ''
      );
    }
    if (type === 'obtenant') {
      const clusterResourceName = (list as API.TenantDetail[]).find(
        (tenant) => {
          return tenant.tenantName === target;
        },
      )?.clusterResourceName;
      if (clusterResourceName) {
        return clusterList?.find(
          (cluster) => cluster.name === clusterResourceName,
        )?.clusterName;
      }
    }
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

  useUpdateEffect(() => {
    const filter: { [T: string]: unknown } = {};
    Object.keys(formData).forEach((key) => {
      if (formData[key]) {
        if (typeof formData[key] === 'string') {
          filter[key] = formData[key];
        } else if (key === 'startTime' || key === 'endTime') {
          filter[key] = Math.ceil(formData[key].valueOf() / 1000);
        } else if (key === 'instance' && formData[key]?.type) {
          const temp = {};
          if (formData[key]?.obtenant) {
            formData[key].obcluster = findClusterName(
              tenantList!,
              'obtenant',
              formData[key]?.obtenant,
            );
          }
          if (formData[key]?.observer) {
            formData[key].obcluster = findClusterName(
              clusterList!,
              'observer',
              formData[key]?.observer,
            );
          }
          Object.keys(formData[key]).forEach((innerKey) => {
            if (formData[key][innerKey]) {
              temp[innerKey] = formData[key][innerKey];
            }
          });
          filter[key] = temp;
        }
      }
    });
    if (filter.instance) {
      if (Object.keys(filter.instance).length === 1 && filter.instance?.type) {
        filter.instanceType = filter.instance.type;
        delete filter.instance;
      }
    }
    debounceDepend(filter);
  }, [formData]);

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
                      allowClear
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
              name={'severity'}
            >
              <Select
                allowClear
                placeholder="请选择"
                options={LEVER_OPTIONS_ALARM!.map((item) => ({
                  value: item.value,
                  label: (
                    <Tag
                      color={
                        SEVERITY_MAP[item.value as Alert.AlarmLevel]?.color
                      }
                    >
                      {item.label}
                    </Tag>
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
