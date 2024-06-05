import type { OceanbaseOBInstanceType } from '@/api/generated';
import { Alert } from '@/type/alert';
import { useModel } from '@umijs/max';
import { useUpdateEffect } from 'ahooks';
import type { FormInstance } from 'antd';
import { Col, Form, Row, Select } from 'antd';
import { flatten } from 'lodash';
import { useEffect, useState } from 'react';
import { getSelectList } from '../helper';

interface ShieldObjInputProps {
  shieldObjType: OceanbaseOBInstanceType;
  form: FormInstance;
}

interface ShieldObjFormItemProps {
  clusterFormName: ['instances', 'obcluster'];
  tenantFormName?: ['instances', 'obtenant'];
  serverFormName?: ['instances', 'observer'];
}

export default function ShieldObjInput({
  shieldObjType,
  form,
}: ShieldObjInputProps) {
  const { clusterList, tenantList } = useModel('alarm');
  const getOptionsFromType = (
    type: OceanbaseOBInstanceType,
    selectedCluster?: string[],
  ) => {
    if (!type || !clusterList || (type === 'obtenant' && !tenantList))
      return [];
    const list = getSelectList(clusterList, type, tenantList);
    if (type === 'obcluster') {
      const res = list?.map((clusterName) => ({
        value: clusterName,
        label: clusterName,
      }));
      if (res.length && shieldObjType === 'obcluster') {
        res.splice(0, 0, { value: 'allClusters', label: '全部集群' });
      }
      return res;
    }
    if (type === 'obtenant') {
      if (selectedCluster?.length) {
        const res = flatten(
          (list as Alert.TenantsList[])?.map((item) =>
            selectedCluster.includes(item.clusterName) ? item.tenants : [],
          ),
        ).map((tenantName) => ({
          label: tenantName,
          value: tenantName,
        }));
        if (res?.length) {
          res.splice(0, 0, { value: 'allTenants', label: '全部租户' });
        }
        return res;
      } else {
        return (list as Alert.TenantsList[])?.map((cluster) => ({
          label: <span>{cluster.clusterName}</span>,
          title: cluster.clusterName,
          options: cluster.tenants?.map((item) => ({
            value: item,
            label: item,
          })),
        }));
      }
    }
    if (type === 'observer') {
      if (selectedCluster?.length) {
        const res = flatten(
          (list as Alert.ServersList[])?.map((item) =>
            selectedCluster.includes(item.clusterName) ? item.servers : [],
          ),
        ).map((server) => ({
          label: server,
          value: server,
        }));
        if (res?.length) {
          res.splice(0, 0, { value: 'allServers', label: '全部主机' });
        }
        return res;
      } else {
        return (list as Alert.ServersList[]).map((cluster) => ({
          label: <span>{cluster.clusterName}</span>,
          title: cluster.clusterName,
          options: cluster.servers?.map((item) => ({
            value: item,
            label: item,
          })),
        }));
      }
    }
  };
  const ShieldObjFormItem = ({
    clusterFormName,
    tenantFormName,
    serverFormName,
  }: ShieldObjFormItemProps) => {
    const nextFormName = tenantFormName || serverFormName;
    const selectedCluster = Form.useWatch(clusterFormName, form);

    const [maxCount, setMaxCount] = useState<number>();
    const selectOnChange = (vals: string[]) => {
      if (
        vals.includes('allTenants') ||
        vals.includes('allServers') ||
        !selectedCluster?.length
      ) {
        setMaxCount(1);
      } else {
        setMaxCount(undefined);
      }
      if (vals.includes('allTenants')) {
        form.setFieldValue(tenantFormName, ['allTenants']);
      }
      if (vals.includes('allServers')) {
        form.setFieldValue(serverFormName, ['allServers']);
      }
    };
    useUpdateEffect(() => {
      if (!selectedCluster?.length) {
        form.setFieldValue(nextFormName, undefined);
      }
    }, [selectedCluster]);
    return (
      <Row gutter={8}>
        <Col span={8}>
          <Form.Item
            rules={[
              {
                required: true,
                message: '请选择',
              },
            ]}
            name={clusterFormName}
          >
            <Select
              mode="multiple"
              maxCount={1}
              allowClear
              options={getOptionsFromType(clusterFormName[1])}
              placeholder="请选择集群"
            />
          </Form.Item>
        </Col>
        <Col span={16}>
          <Form.Item noStyle dependencies={[clusterFormName]}>
            {({ getFieldValue }) => {
              const cluster = getFieldValue(clusterFormName);
              return (
                <Form.Item
                  name={nextFormName}
                  rules={[
                    {
                      required: true,
                      message: '请选择',
                    },
                  ]}
                  dependencies={[clusterFormName]}
                >
                  <Select
                    mode="multiple"
                    onChange={selectOnChange}
                    maxCount={maxCount}
                    allowClear
                    options={getOptionsFromType(nextFormName![1], cluster)}
                    placeholder={`请选择${
                      nextFormName![1] === 'observer' ? '主机' : '租户'
                    }`}
                  />
                </Form.Item>
              );
            }}
          </Form.Item>
        </Col>
      </Row>
    );
  };
  const ShieldObjClusterFormItem = () => {
    const selectedCluster = Form.useWatch(['instances', 'obcluster']);
    const [maxCount, setMaxCount] = useState<number>();
    useEffect(() => {
      if (selectedCluster && selectedCluster.includes('allClusters')) {
        form.setFieldValue(['instances', 'obcluster'], ['allClusters']);
        setMaxCount(1);
      } else {
        setMaxCount(undefined);
      }
    }, [selectedCluster]);
    return (
      <Row>
        <Col span={24}>
          <Form.Item
            rules={[
              {
                required: true,
                message: '请选择',
              },
            ]}
            name={['instances', 'obcluster']}
          >
            <Select
              mode="multiple"
              maxCount={maxCount}
              allowClear
              placeholder="请选择集群"
              options={getOptionsFromType(shieldObjType)}
            />
          </Form.Item>
        </Col>
      </Row>
    );
  };

  return (
    <>
      {shieldObjType === 'obcluster' && <ShieldObjClusterFormItem />}
      {shieldObjType === 'obtenant' && (
        <ShieldObjFormItem
          clusterFormName={['instances', 'obcluster']}
          tenantFormName={['instances', 'obtenant']}
        />
      )}
      {shieldObjType === 'observer' && (
        <ShieldObjFormItem
          clusterFormName={['instances', 'obcluster']}
          serverFormName={['instances', 'observer']}
        />
      )}
    </>
  );
}
