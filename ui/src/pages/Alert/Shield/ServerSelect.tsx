import { Alert } from '@/type/alert';
import { intl } from '@/utils/intl';
import { useModel } from '@umijs/max';
import { useUpdateEffect } from 'ahooks';
import { Col, Row, Select } from 'antd';
import { flatten } from 'lodash';
import { useEffect, useRef, useState } from 'react';
import { getSelectList, isListSame } from '../helper';

type ServerSelectVal = {
  obcluster?: string[];
  observer?: string[];
};

interface ServerSelectProps {
  onChange?: (val: ServerSelectVal) => void;
  value?: ServerSelectVal;
}
export default function ServerSelect({
  onChange,
  value = {},
}: ServerSelectProps) {
  const [maxCount, setMaxCount] = useState<number | undefined>(undefined);
  const { clusterList = [] } = useModel('alarm');
  const [clusterOptions, setClusterOptions] = useState<API.OptionsType>([]);
  const preOBServersRef = useRef(value.observer);
  const [serversOptions, setServersOptions] = useState<any>([]);

  const clusterChange = (selectedCluster: string[]) => {
    onChange?.({ ...value, obcluster: selectedCluster });
  };

  const serverChange = (selectedServers: string[]) => {
    onChange?.({ ...value, observer: selectedServers });
  };

  const getServerDefaultOptions = () => {
    return (getSelectList(clusterList, 'observer') as Alert.ServersList[]).map(
      (cluster) => ({
        label: <span>{cluster.clusterName}</span>,
        title: cluster.clusterName,
        options: cluster.servers?.map((item) => ({
          value: item,
          label: item,
        })),
      }),
    );
  };
  const getServersOfClusterOptions = (selectedServers: string[]) => {
    const servers = getSelectList(clusterList, 'observer');
    const res = flatten(
      (servers as Alert.ServersList[])?.map((item) =>
        selectedServers.includes(item.clusterName) ? item.servers : [],
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
  };
  const getClusterDefaultOptions = () => {
    return (getSelectList(clusterList, 'obcluster') as string[]).map(
      (item) => ({
        label: item,
        value: item,
      }),
    );
  };
  const getClusterFromServersOptions = (selectServers: string[]) => {
    const list = getSelectList(
      clusterList,
      'obcluster',
      undefined,
      undefined,
      selectServers,
    );
    return list?.map((clusterName) => ({
      value: clusterName,
      label: clusterName,
    }));
  };
  useEffect(() => {
    if (clusterList?.length) {
      setClusterOptions(getClusterDefaultOptions());
      setServersOptions(getServerDefaultOptions());
    }
  }, []);

  useUpdateEffect(() => {
    // clear obcluster
    if (!value.obcluster?.length) {
      setServersOptions(getServerDefaultOptions());
      onChange?.({ ...value, observer: [] });
    } else {
      setServersOptions(getServersOfClusterOptions(value.obcluster));
    }
  }, [value.obcluster]);

  useUpdateEffect(() => {
    //clear observers
    if (!value.observer?.length) {
      setClusterOptions(getClusterDefaultOptions());
    } else {
      if (value.observer.includes('allServers')) {
        if (!isListSame(preOBServersRef.current || [], value.observer)) {
          onChange?.({ ...value, observer: ['allServers'] });
          setMaxCount(1);
        }
      } else {
        maxCount === 1 && setMaxCount(undefined);
        const newClusterOptions = getClusterFromServersOptions(
          value.observer,
        ) as API.OptionsType;
        onChange?.({ ...value, obcluster: [newClusterOptions[0]?.value] });
        setClusterOptions(newClusterOptions);
      }
    }
    preOBServersRef.current = value.observer;
  }, [value.observer]);

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
          value={value.observer}
          maxCount={maxCount}
          onChange={serverChange}
          options={serversOptions}
          allowClear
          placeholder={intl.formatMessage({
            id: 'src.pages.Alert.Shield.E7B0A331',
            defaultMessage: '请选择主机',
          })}
        />
      </Col>
    </Row>
  );
}
