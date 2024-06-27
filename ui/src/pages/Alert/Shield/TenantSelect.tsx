import { Alert } from '@/type/alert';
import { intl } from '@/utils/intl';
import { useModel } from '@umijs/max';
import { useUpdateEffect } from 'ahooks';
import { Col, Row, Select } from 'antd';
import { flatten } from 'lodash';
import { useEffect, useRef, useState } from 'react';
import { getSelectList, isListSame } from '../helper';

type TenantSelectVal = {
  obcluster?: string[];
  obtenant?: string[];
};

interface TenantSelectProps {
  onChange?: (val: TenantSelectVal) => void;
  value?: TenantSelectVal;
}
export default function TenantSelect({
  onChange,
  value = {},
}: TenantSelectProps) {
  const [maxCount, setMaxCount] = useState<number | undefined>(undefined);
  const { clusterList = [], tenantList = [] } = useModel('alarm');
  const [clusterOptions, setClusterOptions] = useState<API.OptionsType>([]);
  const preOBTenantRef = useRef(value.obtenant);
  const [tenantOptions, setTenantOptions] = useState<any>([]);

  const clusterChange = (selectedCluster: string[]) => {
    onChange?.({ ...value, obcluster: selectedCluster });
  };

  const tenantChange = (selectedTenant: string[]) => {
    onChange?.({ ...value, obtenant: selectedTenant });
  };

  const getTenantDefaultOptions = () => {
    return (
      getSelectList(clusterList, 'obtenant', tenantList) as Alert.TenantsList[]
    ).map((cluster) => ({
      label: <span>{cluster.clusterName}</span>,
      title: cluster.clusterName,
      options: cluster.tenants?.map((item) => ({
        value: item,
        label: item.includes('sys') ? 'sys' : item,
      })),
    }));
  };
  const getTenantsOfClusterOptions = (selectedCluster: string[]) => {
    const tenants = getSelectList(clusterList, 'obtenant', tenantList);
    const res = flatten(
      (tenants as Alert.TenantsList[])?.map((item) =>
        selectedCluster.includes(item.clusterName) ? item.tenants : [],
      ),
    ).map((tenantName) => ({
      label: tenantName?.includes('sys') ? 'sys' : tenantName,
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
  };
  const getClusterDefaultOptions = () => {
    return (getSelectList(clusterList, 'obcluster') as string[]).map(
      (item) => ({
        label: item,
        value: item,
      }),
    );
  };
  const getClusterFromTenantOptions = (selectTenant: string[]) => {
    const list = getSelectList(
      clusterList,
      'obcluster',
      tenantList,
      selectTenant,
    );
    return list?.map((clusterName) => ({
      value: clusterName,
      label: clusterName,
    }));
  };
  useEffect(() => {
    if (clusterList?.length && tenantList?.length) {
      if (value.obtenant?.length) {
        setClusterOptions(getClusterFromTenantOptions(value.obtenant));
      } else {
        setClusterOptions(getClusterDefaultOptions());
      }
      if (value.obcluster?.length) {
        setTenantOptions(getTenantsOfClusterOptions(value.obcluster));
      } else {
        setTenantOptions(getTenantDefaultOptions());
      }
    }
  }, []);

  useUpdateEffect(() => {
    // clear obcluster
    if (!value.obcluster?.length) {
      setTenantOptions(getTenantDefaultOptions());
      onChange?.({ ...value, obtenant: [] });
    } else {
      setTenantOptions(getTenantsOfClusterOptions(value.obcluster));
    }
  }, [value.obcluster]);

  useUpdateEffect(() => {
    //clear obtenant
    if (!value.obtenant?.length) {
      setClusterOptions(getClusterDefaultOptions());
    } else {
      if (value.obtenant.includes('allTenants')) {
        if (!isListSame(preOBTenantRef.current || [], value.obtenant)) {
          onChange?.({ ...value, obtenant: ['allTenants'] });
          setMaxCount(1);
        }
      } else {
        maxCount === 1 && setMaxCount(undefined);
        const newClusterOptions = getClusterFromTenantOptions(
          value.obtenant,
        ) as API.OptionsType;
        onChange?.({ ...value, obcluster: [newClusterOptions[0]?.value] });
        setClusterOptions(newClusterOptions);
      }
    }
    preOBTenantRef.current = value.obtenant;
  }, [value.obtenant]);

  return (
    <Row gutter={8}>
      <Col span={8}>
        <Select
          mode="multiple"
          maxCount={1}
          value={value.obcluster}
          allowClear
          options={clusterOptions}
          onChange={clusterChange}
          placeholder={intl.formatMessage({
            id: 'src.pages.Alert.Shield.B0133BD9',
            defaultMessage: '请选择集群',
          })}
        />
      </Col>
      <Col span={16}>
        <Select
          mode="multiple"
          value={value.obtenant}
          maxCount={maxCount}
          onChange={tenantChange}
          options={tenantOptions}
          allowClear
          placeholder="请选择租户"
        />
      </Col>
    </Row>
  );
}
