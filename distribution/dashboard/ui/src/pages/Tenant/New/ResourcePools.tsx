import { SUFFIX_UNIT } from '@/constants';
import { intl } from '@/utils/intl';
import { Card, Checkbox, Col, Form, InputNumber, Row } from 'antd';
import { FormInstance } from 'antd/lib/form';
import { useEffect, useState } from 'react';
import styles from './index.less';

interface ResourcePoolsProps {
  selectClusterId?: number;
  clusterList: API.SimpleClusterList;
  form: FormInstance<any>;
}

export default function ResourcePools({
  selectClusterId,
  clusterList,
  form,
}: ResourcePoolsProps) {
  const ZoneItem = ({ name }: { name: string }) => {
    const [isChecked, setIsChecked] = useState<boolean>(false);

    useEffect(() => {
      if (!isChecked)
        form.setFieldValue(['pools', name, 'priority'], undefined);
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
          value={isChecked}
          style={{ marginRight: 24 }}
          onChange={(e) => setIsChecked(!e.target.value)}
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
      </div>
    );
  };

  const targetZoneList = clusterList
    .filter((cluster) => cluster.clusterId === selectClusterId)[0]
    ?.topology.map((zone) => ({ zone: zone.name }));
  return (
    <Card
      title={intl.formatMessage({
        id: 'Dashboard.Tenant.New.ResourcePools.ResourcePool',
        defaultMessage: '资源池',
      })}
    >
      <Row>
        <p>Unit config</p>
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
                label={intl.formatMessage({
                  id: 'Dashboard.Tenant.New.ResourcePools.LogDisk',
                  defaultMessage: '日志磁盘',
                })}
              >
                <InputNumber
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
                <InputNumber placeholder="min" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="max iops" name={['unitConfig', 'maxIops']}>
                <InputNumber placeholder="max" style={{ width: '100%' }} />
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
      </Row>
      <Row>
        <h1>
          {intl.formatMessage({
            id: 'Dashboard.Tenant.New.ResourcePools.ZonePriority',
            defaultMessage: 'Zone优先级',
          })}
        </h1>

        {targetZoneList &&
          targetZoneList.map((item, index) => (
            <ZoneItem key={index} name={item.zone} />
          ))}
      </Row>
    </Card>
  );
}
