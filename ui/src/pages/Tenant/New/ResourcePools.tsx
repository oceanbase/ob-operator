import InputNumber from '@/components/InputNumber';
import { SUFFIX_UNIT } from '@/constants';
import { intl } from '@/utils/intl';
import { Card, Checkbox, Col, Form, Row, Tooltip } from 'antd';
import { FormInstance } from 'antd/lib/form';
import { useEffect, useMemo, useState } from 'react';
import { findMinParameter } from '../helper';
import styles from './index.less';

interface ResourcePoolsProps {
  selectClusterId?: number;
  clusterList: API.SimpleClusterList;
  form: FormInstance<any>;
  essentialParameter?: API.EssentialParametersType;
}

export type MaxResourceType = {
  maxCPU?: number;
  maxLogDisk?: number;
  maxMemory?: number;
};

export default function ResourcePools({
  selectClusterId,
  clusterList,
  essentialParameter,
  form,
}: ResourcePoolsProps) {
  const [minMemory, setMinMemory] = useState<number>(2);
  const [maxResource, setMaxResource] = useState<MaxResourceType>({});
  const [selectZones, setSelectZones] = useState<string[]>([]);
  const ZoneItem = ({ name }: { name: string }) => {
    const [isChecked, setIsChecked] = useState<boolean>(false);
    const obZoneResource = essentialParameter?.obZoneResourceMap[name];
    useEffect(() => {
      if (!isChecked) {
        form.setFieldValue(['pools', name, 'priority'], undefined);
        setSelectZones(selectZones.filter((zone) => zone !== name));
      } else {
        setSelectZones([...selectZones, name]);
      }
    }, [isChecked]);
    return (
      <div
        style={{
          width: '100%',
          display: 'flex',
          justifyContent: 'flex-start',
          alignItems: 'center',
        }}
      >
        <span style={{ marginRight: 8 }}>{name}</span>
        <Checkbox
          checked={isChecked}
          style={{ marginRight: 24 }}
          onChange={(e) => {
            setIsChecked(e.target.checked);
          }}
        />

        <Col span={8}>
          <Form.Item
            name={['pools', name, 'priority']}
            label={intl.formatMessage({
              id: 'Dashboard.Tenant.New.ResourcePools.Weight',
              defaultMessage: '权重',
            })}
          >
            <InputNumber style={{ width: '100%' }} disabled={!isChecked} />
          </Form.Item>
        </Col>
        {obZoneResource && (
          <Col style={{ marginLeft: 12 }} span={8}>
            <span style={{ marginRight: 12 }}>
              {intl.formatMessage({
                id: 'Dashboard.Tenant.New.ResourcePools.AvailableResources',
                defaultMessage: '可用资源：',
              })}
            </span>
            <span style={{ marginRight: 12 }}>
              CPU {obZoneResource['availableCPU']}
            </span>
            <span style={{ marginRight: 12 }}>
              Memory {obZoneResource['availableMemory'] / (1 << 30)}GB
            </span>
            <span>
              {intl.formatMessage({
                id: 'Dashboard.Tenant.New.ResourcePools.LogDiskSize',
                defaultMessage: '日志磁盘大小',
              })}
              {obZoneResource['availableLogDisk'] / (1 << 30)}GB
            </span>
          </Col>
        )}
      </div>
    );
  };
  const targetZoneList = clusterList
    .filter((cluster) => cluster.clusterId === selectClusterId)[0]
    ?.topology.map((zone) => ({ zone: zone.zone }));

  useEffect(() => {
    if (essentialParameter) {
      setMinMemory(essentialParameter.minPoolMemory / (1 << 30));
    }
  }, [essentialParameter]);

  useEffect(() => {
    if (essentialParameter) {
      if (selectZones.length === 0) {
        setMaxResource({});
        return;
      }
      setMaxResource(findMinParameter(selectZones, essentialParameter));
    }
  }, [selectZones]);

  return (
    <Card
      title={intl.formatMessage({
        id: 'Dashboard.Tenant.New.ResourcePools.ResourcePool',
        defaultMessage: '资源池',
      })}
    >
      <div>
        {useMemo(
          () => (
            <div>
              {' '}
              {targetZoneList && essentialParameter && (
                <Row>
                  <h3>
                    {intl.formatMessage({
                      id: 'Dashboard.Tenant.New.ResourcePools.ZonePriority',
                      defaultMessage: 'Zone优先级',
                    })}
                  </h3>
                  {targetZoneList.map((item, index) => (
                    <ZoneItem key={index} name={item.zone} />
                  ))}
                </Row>
              )}
            </div>
          ),

          [selectClusterId, clusterList, essentialParameter],
        )}

        <h3>Unit config</h3>
        <div className={styles.unitConfigContainer}>
          <Row gutter={[16, 32]}>
            <Col span={8}>
              <Form.Item
                name={['unitConfig', 'cpuCount']}
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.Tenant.New.ResourcePools.PleaseEnterCpuCore',
                      defaultMessage: '请输入 CPU (核)',
                    }),
                  },
                ]}
                label="CPU"
              >
                <InputNumber
                  addonAfter={
                    <div>
                      {intl.formatMessage({
                        id: 'Dashboard.Tenant.New.ResourcePools.Nuclear',
                        defaultMessage: '核',
                      })}
                    </div>
                  }
                  min={1}
                  max={maxResource.maxCPU}
                  placeholder={intl.formatMessage({
                    id: 'Dashboard.Tenant.New.ResourcePools.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name={['unitConfig', 'memorySize']}
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.Tenant.New.ResourcePools.EnterMemorySize',
                      defaultMessage: '请输入Memory size',
                    }),
                  },
                ]}
                label="Memory"
              >
                <InputNumber
                  addonAfter={SUFFIX_UNIT}
                  min={minMemory}
                  max={
                    maxResource.maxMemory
                      ? maxResource.maxMemory / (1 << 30)
                      : undefined
                  }
                  placeholder={intl.formatMessage({
                    id: 'Dashboard.Tenant.New.ResourcePools.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name={['unitConfig', 'logDiskSize']}
                label={
                  <Tooltip
                    title={intl.formatMessage({
                      id: 'Dashboard.Tenant.New.ResourcePools.ThisRefersToTheTenant',
                      defaultMessage: '此处指的是租户的 clog 磁盘空间',
                    })}
                  >
                    {intl.formatMessage({
                      id: 'Dashboard.Tenant.New.ResourcePools.LogDiskSize',
                      defaultMessage: '日志磁盘大小',
                    })}
                  </Tooltip>
                }
              >
                <InputNumber
                  min={5}
                  max={
                    maxResource.maxLogDisk
                      ? maxResource.maxLogDisk / (1 << 30)
                      : undefined
                  }
                  addonAfter={SUFFIX_UNIT}
                  placeholder={intl.formatMessage({
                    id: 'Dashboard.Tenant.New.ResourcePools.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="min iops" name={['unitConfig', 'minIops']}>
                <InputNumber
                  min={1024}
                  placeholder="min"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="max iops" name={['unitConfig', 'maxIops']}>
                <InputNumber
                  min={1024}
                  placeholder="max"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                label={intl.formatMessage({
                  id: 'Dashboard.Tenant.New.ResourcePools.IopsWeight',
                  defaultMessage: 'iops 权重',
                })}
                name={['unitConfig', 'iopsWeight']}
              >
                <InputNumber
                  placeholder={intl.formatMessage({
                    id: 'Dashboard.Tenant.New.ResourcePools.PleaseEnter',
                    defaultMessage: '请输入',
                  })}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>
        </div>
      </div>
    </Card>
  );
}
