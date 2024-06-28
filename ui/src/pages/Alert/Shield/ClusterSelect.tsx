import { intl } from '@/utils/intl';
import { useModel } from '@umijs/max';
import { Col, Row, Select } from 'antd';
import { useEffect, useRef, useState } from 'react';
import { getSelectList, isListSame } from '../helper';

type ClusterType = {
  obcluster?: string[];
};

interface ClusterSelectProps {
  onChange?: (value: ClusterType) => void;
  value?: ClusterType;
}

export default function ClusterSelect({
  onChange,
  value = {},
}: ClusterSelectProps) {
  const { clusterList = [] } = useModel('alarm');
  const [maxCount, setMaxCount] = useState<number>();
  const [clusterOptions, setClusterOptions] = useState<API.OptionsType>([]);
  const preOBClusterRef = useRef(value.obcluster);

  const getClusterDefaultOptions = () => {
    const list = getSelectList(clusterList, 'obcluster');
    const res = list?.map((clusterName) => ({
      value: clusterName,
      label: clusterName,
    }));
    if (res.length) {
      res.splice(0, 0, {
        value: 'allClusters',
        label: intl.formatMessage({
          id: 'src.pages.Alert.Shield.E34008B6',
          defaultMessage: '全部集群',
        }),
      });
    }
    return res;
  };
  const clusterChange = (selectedCluster: string[]) => {
    onChange?.({ ...value, obcluster: selectedCluster });
  };
  useEffect(() => {
    if (value?.obcluster?.length && value.obcluster.includes('allClusters')) {
      if (!isListSame(preOBClusterRef.current || [], value.obcluster)) {
        onChange?.({ ...value, obcluster: ['allClusters'] });
        setMaxCount(1);
      }
    } else {
      maxCount === 1 && setMaxCount(undefined);
      setClusterOptions(getClusterDefaultOptions());
      setMaxCount(undefined);
    }
    preOBClusterRef.current = value.obcluster;
  }, [value.obcluster]);
  return (
    <Row>
      <Col span={24}>
        <Select
          mode="multiple"
          maxCount={maxCount}
          value={value.obcluster}
          allowClear
          options={clusterOptions}
          onChange={clusterChange}
          placeholder={intl.formatMessage({
            id: 'src.pages.Alert.Shield.2E5B9DD9',
            defaultMessage: '请选择集群',
          })}
        />
      </Col>
    </Row>
  );
}
