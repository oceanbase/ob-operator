import InputNumber from '@/components/InputNumber';
import { MAX_IOPS, SUFFIX_UNIT, getMinResource } from '@/constants';
import { getClusterDetailReq } from '@/services';
import { intl } from '@/utils/intl';
import { useRequest } from 'ahooks';
import { Card, Col, Form, Row, Tooltip } from 'antd';
import { FormInstance } from 'antd/lib/form';
import { useEffect, useState } from 'react';
import ZoneItem from '../ZoneItem';
import { findMinParameter, modifyZoneCheckedStatus } from '../helper';
import styles from './index.less';

interface ResourcePoolsProps {
  type?: string;
  obVersion?: string;
  selectClusterId?: string;
  clusterList: API.SimpleClusterList;
  form: FormInstance<API.NewTenantForm>;
  setClusterList: React.Dispatch<React.SetStateAction<API.SimpleClusterList>>;
  essentialParameter?: API.EssentialParametersType;
}

export default function ResourcePools({
  selectClusterId,
  clusterList,
  essentialParameter,
  setClusterList,
  form,
  type,
}: ResourcePoolsProps) {
  const [minResource, setMinResource] = useState<OBTenant.MinResourceConfig>(
    getMinResource(),
  );
  const [maxResource, setMaxResource] = useState<OBTenant.MaxResourceType>({});
  const [selectZones, setSelectZones] = useState<string[]>([]);
  const [obversion, setObversion] = useState<string>('');

  const checkBoxOnChange = (checked: boolean, name: string) => {
    form.setFieldValue(['pools', name, 'checked'], checked);
    if (!checked) {
      form.setFieldValue(['pools', name, 'priority'], undefined);
      setSelectZones(selectZones.filter((zone) => zone !== name));
    } else {
      // form.setFieldValue(['pools',name])
      setSelectZones([...selectZones, name]);
    }
    setClusterList(
      modifyZoneCheckedStatus(clusterList, name, checked, {
        id: selectClusterId,
      }),
    );
  };
  const targetZoneList = clusterList
    .filter((cluster) => cluster.id === selectClusterId)[0]
    ?.topology.map((zone) => ({ zone: zone.zone, checked: zone.checked }));

  const { run: getClusterDetail } = useRequest(getClusterDetailReq, {
    manual: true,
    onSuccess: (res) => {
      if (res?.info?.version) {
        setObversion(res?.info?.version);
      }
    },
  });

  useEffect(() => {
    if (selectClusterId) {
      const dp = clusterList
        .filter((cluster) => cluster.id === selectClusterId)
        ?.map((item) => ({ name: item.name, ns: item.namespace }));
      getClusterDetail(dp[0]);
    }
  }, [selectClusterId]);

  useEffect(() => {
    if (essentialParameter) {
      if (selectZones.length === 0) {
        setMaxResource({});
        return;
      }
      const maxResource = findMinParameter(selectZones, essentialParameter);
      if (maxResource.maxCPU! < minResource.minCPU) {
        maxResource.maxCPU = minResource.minCPU;
      }
      if (maxResource.maxLogDisk! < minResource.minLogDisk) {
        maxResource.maxLogDisk = minResource.minLogDisk;
      }
      if (maxResource.maxMemory! < minResource.minMemory) {
        maxResource.maxMemory = minResource.minMemory;
      }
      setMaxResource(maxResource);
    }
  }, [selectZones, essentialParameter]);

  useEffect(() => {
    if (essentialParameter) {
      setMinResource(
        getMinResource({
          minMemory: essentialParameter?.minPoolMemory,
        }),
      );
    }
  }, [essentialParameter]);

  const content = () => {
    return (
      <div>
        {targetZoneList && essentialParameter && (
          <Row>
            <h3>副本发布</h3>
            {targetZoneList.map((item, index) => (
              <ZoneItem
                key={index}
                type={type || 'new'}
                name={item.zone}
                checked={item.checked!}
                checkedFormName={['pools', item.zone, 'checked']}
                obZoneResource={
                  essentialParameter?.obZoneResourceMap[item.zone]
                }
                obVersion={obversion}
                checkBoxOnChange={checkBoxOnChange}
              />
            ))}
            <Col span={8}>
              <Form.Item
                name={['unitNum']}
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.Tenant.New.BasicInfo.PleaseEnterTheNumberOf',
                      defaultMessage: '请输入Unit 数量',
                    }),
                  },
                ]}
                label={'每 zone unit 数量'}
              >
                <InputNumber min={1} />
              </Form.Item>
            </Col>
          </Row>
        )}

        <h3>
          {intl.formatMessage({
            id: 'Dashboard.Tenant.New.ResourcePools.ResourceUnitSpecifications',
            defaultMessage: '资源单元规格',
          })}
        </h3>
        <div className={styles.unitConfigContainer}>
          <Row gutter={[16, 32]}>
            <Col span={8}>
              <Form.Item
                name={['unitConfig', 'cpuCount']}
                dependencies={targetZoneList?.map((zone) => [
                  'pools',
                  zone.zone,
                  'checked',
                ])}
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.Tenant.New.ResourcePools.PleaseEnterCpuCore',
                      defaultMessage: '请输入 CPU (核)',
                    }),
                  },
                  () => ({
                    validator() {
                      if (
                        essentialParameter &&
                        selectZones.length &&
                        findMinParameter(selectZones, essentialParameter)
                          .maxCPU! < minResource.minCPU
                      ) {
                        return Promise.reject(
                          new Error(
                            intl.formatMessage({
                              id: 'Dashboard.Tenant.New.ResourcePools.ZoneCannotCreateAUnit',
                              defaultMessage:
                                '可用 CPU 过小， Zone 无法创建 Unit',
                            }),
                          ),
                        );
                      }
                      return Promise.resolve();
                    },
                  }),
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
                  min={minResource.minCPU}
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
                dependencies={targetZoneList?.map((zone) => [
                  'pools',
                  zone.zone,
                  'checked',
                ])}
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.Tenant.New.ResourcePools.EnterMemorySize',
                      defaultMessage: '请输入 Memory size',
                    }),
                  },
                  () => ({
                    validator() {
                      if (
                        essentialParameter &&
                        selectZones.length &&
                        findMinParameter(selectZones, essentialParameter)
                          .maxMemory! < minResource.minMemory
                      ) {
                        return Promise.reject(
                          new Error(
                            intl.formatMessage({
                              id: 'Dashboard.Tenant.New.ResourcePools.IfTheAvailableMemorySize',
                              defaultMessage:
                                '可用 Memory size 过小，Zone 将无法创建 Unit',
                            }),
                          ),
                        );
                      }
                      return Promise.resolve();
                    },
                  }),
                ]}
                label="Memory"
              >
                <InputNumber
                  addonAfter={SUFFIX_UNIT}
                  min={minResource.minMemory}
                  max={
                    maxResource.maxMemory ? maxResource.maxMemory : undefined
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
                dependencies={targetZoneList?.map((zone) => [
                  'pools',
                  zone.zone,
                  'checked',
                ])}
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.Tenant.New.ResourcePools.EnterTheLogDiskSize',
                      defaultMessage: '请输入日志磁盘大小',
                    }),
                  },
                  () => ({
                    validator() {
                      if (
                        essentialParameter &&
                        selectZones.length &&
                        findMinParameter(selectZones, essentialParameter)
                          .maxLogDisk! < minResource.minLogDisk
                      ) {
                        return Promise.reject(
                          new Error(
                            intl.formatMessage({
                              id: 'Dashboard.Tenant.New.ResourcePools.ZoneCannotCreateAUnit.1',
                              defaultMessage:
                                '可用日志磁盘空间过小， Zone 无法创建 Unit',
                            }),
                          ),
                        );
                      }
                      return Promise.resolve();
                    },
                  }),
                ]}
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
                  min={minResource.minLogDisk}
                  max={
                    maxResource.maxLogDisk ? maxResource.maxLogDisk : undefined
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
                  min={minResource.minIops}
                  max={MAX_IOPS}
                  placeholder="min"
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="max iops" name={['unitConfig', 'maxIops']}>
                <InputNumber
                  min={minResource.maxIops}
                  max={MAX_IOPS}
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
    );
  };
  return (
    <>
      {type === 'tenantBackup' ? (
        <> {content()}</>
      ) : (
        <Card
          title={intl.formatMessage({
            id: 'Dashboard.Tenant.New.ResourcePools.ResourcePool',
            defaultMessage: '资源池',
          })}
        >
          {content()}
        </Card>
      )}
    </>
  );
}
