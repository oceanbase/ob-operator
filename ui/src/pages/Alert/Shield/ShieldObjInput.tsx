import type { OceanbaseOBInstanceType } from '@/api/generated';
import { Alert } from '@/type/alert';
import { intl } from '@/utils/intl';
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
    const selectedTenants = form.getFieldValue(['instances', 'obtenant']);
    const selectedServers = form.getFieldValue(['instances', 'observer']);
    const list = getSelectList(
      clusterList,
      type,
      tenantList,
      selectedTenants,
      selectedServers,
    );
    if (type === 'obcluster') {
      const res = list?.map((clusterName) => ({
        value: clusterName,
        label: clusterName,
      }));
      if (res.length && shieldObjType === 'obcluster') {
        res.splice(0, 0, {
          value: 'allClusters',
          label: intl.formatMessage({
            id: 'src.pages.Alert.Shield.E34008B6',
            defaultMessage: '全部集群',
          }),
        });
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
          res.splice(0, 0, {
            value: 'allTenants',
            label: intl.formatMessage({
              id: 'src.pages.Alert.Shield.DC28F74E',
              defaultMessage: '全部租户',
            }),
          });
        }
        return res;
      } else {
        return (list as Alert.TenantsList[])?.map((cluster) => ({
          label: <span>{cluster.clusterName}</span>,
          title: cluster.clusterName,
          options: cluster.tenants?.map((item) => ({
            value: item,
            disabled: selectedTenants?.length
              ? !cluster.tenants?.includes(selectedTenants[0])
              : false,
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
          res.splice(0, 0, {
            value: 'allServers',
            label: intl.formatMessage({
              id: 'src.pages.Alert.Shield.5EFCC526',
              defaultMessage: '全部主机',
            }),
          });
        }
        return res;
      } else {
        return (list as Alert.ServersList[]).map((cluster) => ({
          label: <span>{cluster.clusterName}</span>,
          title: cluster.clusterName,
          options: cluster.servers?.map((item) => ({
            value: item,
            disabled: selectedServers?.length
              ? !cluster.servers?.includes(selectedServers[0])
              : false,
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
      if (vals.includes('allTenants') || vals.includes('allServers')) {
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
        form.setFieldValue(clusterFormName, undefined);
      }
    }, [selectedCluster]);
    return (
      <Row gutter={8}>
        <Col span={8}>
          <Form.Item noStyle dependencies={[nextFormName]}>
            {() => (
              <Form.Item
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'src.pages.Alert.Shield.639D1A8C',
                      defaultMessage: '请选择',
                    }),
                  },
                ]}
                name={clusterFormName}
              >
                <Select
                  mode="multiple"
                  maxCount={1}
                  allowClear
                  options={getOptionsFromType(clusterFormName[1])}
                  placeholder={intl.formatMessage({
                    id: 'src.pages.Alert.Shield.B0133BD9',
                    defaultMessage: '请选择集群',
                  })}
                />
              </Form.Item>
            )}
          </Form.Item>
        </Col>
        <Col span={16}>
          <Form.Item noStyle dependencies={[clusterFormName, nextFormName]}>
            {({ getFieldValue }) => {
              const cluster = getFieldValue(clusterFormName);
              return (
                <Form.Item
                  name={nextFormName}
                  rules={[
                    {
                      required: true,
                      message: intl.formatMessage({
                        id: 'src.pages.Alert.Shield.A519C403',
                        defaultMessage: '请选择',
                      }),
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
                    placeholder={intl.formatMessage(
                      {
                        id: 'src.pages.Alert.Shield.4AE1863A',
                        defaultMessage: '请选择{ConditionalExpression0}',
                      },
                      {
                        ConditionalExpression0:
                          nextFormName![1] === 'observer'
                            ? intl.formatMessage({
                                id: 'src.pages.Alert.Shield.0EEFD182',
                                defaultMessage: '主机',
                              })
                            : intl.formatMessage({
                                id: 'src.pages.Alert.Shield.74A00B6E',
                                defaultMessage: '租户',
                              }),
                      },
                    )}
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
                message: intl.formatMessage({
                  id: 'src.pages.Alert.Shield.5F7B1190',
                  defaultMessage: '请选择',
                }),
              },
            ]}
            name={['instances', 'obcluster']}
          >
            <Select
              mode="multiple"
              maxCount={maxCount}
              allowClear
              placeholder={intl.formatMessage({
                id: 'src.pages.Alert.Shield.2E5B9DD9',
                defaultMessage: '请选择集群',
              })}
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
