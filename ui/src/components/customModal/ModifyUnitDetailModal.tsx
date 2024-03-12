import { SUFFIX_UNIT } from '@/constants';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import type { MaxResourceType } from '@/pages/Tenant/New/ResourcePools';
import ZoneItem from '@/pages/Tenant/ZoneItem';
import { findMinParameter,getNewClusterList } from '@/pages/Tenant/helper';
import { patchTenantConfiguration } from '@/services/tenant';
import { formatUnitDetailData } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { useEffect,useState } from 'react';
import InputNumber from '@/components/InputNumber';

import { Col, Form, Row, message } from 'antd';
import type { CommonModalType } from '.';
import CustomModal from '.';

export type UnitDetailType = {
  unitConfig: {
    unitConfig: {
      cpuCount: number|string;
      iopsWeight: number;
      logDiskSize: number | string;
      maxIops: number;
      memorySize: number | string;
      minIops: number;
    };
    pools:any;
  };
};

type UnitConfigType = {
  clusterList?: API.SimpleClusterList;
  clusterName?: string;
  essentialParameter?: API.EssentialParametersType;
  setClusterList: React.Dispatch<React.SetStateAction<API.SimpleClusterList>>;
};

export default function ModifyUnitDetailModal({
  visible,
  setVisible,
  successCallback,
  clusterList = [],
  setClusterList,
  essentialParameter = {},
  clusterName = '',
}: CommonModalType & UnitConfigType) {
  const [form] = Form.useForm<UnitDetailType>();
  const [minMemory, setMinMemory] = useState<number>(2);
  const [maxResource, setMaxResource] = useState<MaxResourceType>({});
  const [selectZones, setSelectZones] = useState<string[]>([]);
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };
  const handleCancel = () => setVisible(false);
  const onFinish = async (values: any) => {
    const [ns, name] = getNSName();
    const res = await patchTenantConfiguration({
      ns,
      name,
      ...formatUnitDetailData(values),
    });
    if (res.successful) {
      message.success(res.message);
      successCallback();
      form.resetFields();
      setVisible(false);
    }
  };
  const checkBoxOnChange = (checked: boolean, name: string) => {
    if (!checked) {
      form.setFieldValue(['pools', name, 'priority'], undefined);
      setSelectZones(selectZones.filter((zone) => zone !== name));
    } else {
      setSelectZones([...selectZones, name]);
    }
    setClusterList(
      getNewClusterList(clusterList, name, checked, { name: clusterName }),
    );
  };

  const targetZoneList = clusterList
    .filter((cluster) => cluster.clusterName === clusterName)[0]
    ?.topology.map((zone) => ({ zone: zone.zone, checked: zone.checked }));

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
    <CustomModal
      width={780}
      title={intl.formatMessage({
        id: 'Dashboard.components.customModal.ModifyUnitDetailModal.AdjustUnitSpecifications',
        defaultMessage: '调整 Unit 规格',
      })}
      isOpen={visible}
      handleOk={handleSubmit}
      handleCancel={handleCancel}
    >
      <Form
        form={form}
        onFinish={onFinish}
        layout="vertical"
        autoComplete="off"
      >
        {targetZoneList && essentialParameter && (
          <Row>
            <h3>
              {intl.formatMessage({
                id: 'Dashboard.Tenant.New.ResourcePools.ZonePriority',
                defaultMessage: 'Zone优先级',
              })}
            </h3>
            {targetZoneList.map((item, index) => (
              <ZoneItem
                key={index}
                name={item.zone}
                formName={['unitConfig', 'pools', item.zone, 'priority']}
                checked={item.checked!}
                obZoneResource={essentialParameter.obZoneResourceMap[item.zone]}
                checkBoxOnChange={checkBoxOnChange}
              />
            ))}
          </Row>
        )}
        <Row>
          <Col span={8}>
            <Form.Item
              label="CPU"
              name={['unitConfig', 'unitConfig', 'cpuCount']}
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EnterTheNumberOfCpu',
                    defaultMessage: '请输入 CPU 核数',
                  }),
                },
              ]}
            >
              <InputNumber
                min={1}
                max={maxResource.maxCPU}
                addonAfter={intl.formatMessage({
                  id: 'Dashboard.components.customModal.ModifyUnitDetailModal.Nuclear',
                  defaultMessage: '核',
                })}
                placeholder={intl.formatMessage({
                  id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              label="Memory"
              name={['unitConfig', 'unitConfig', 'memorySize']}
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EnterMemory',
                    defaultMessage: '请输入 Memory',
                  }),
                },
              ]}
            >
              <InputNumber
                min={minMemory}
                max={
                  maxResource.maxMemory
                    ? maxResource.maxMemory / (1 << 30)
                    : undefined
                }
                addonAfter={SUFFIX_UNIT}
                placeholder={intl.formatMessage({
                  id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Form.Item
              label="LogDiskSize"
              name={['unitConfig', 'unitConfig', 'logDiskSize']}
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
                  id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
          <Col span={8}>
            <Row gutter={24}>
              <Col>
                <Form.Item
                  label="min iops" 
                  name={['unitConfig', 'unitConfig', 'minIops']}
                >
                  <InputNumber
                     min={1024}
                    placeholder={intl.formatMessage({
                      id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
                      defaultMessage: '请输入',
                    })}
                  />
                </Form.Item>
              </Col>
              <Col>
                <Form.Item
                  label="max iops"
                  name={['unitConfig', 'unitConfig', 'maxIops']}
                >
                  <InputNumber
                    min={1024}
                    placeholder={intl.formatMessage({
                      id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
                      defaultMessage: '请输入',
                    })}
                  />
                </Form.Item>
              </Col>
            </Row>
          </Col>
          <Col span={8}>
            <Form.Item
              label={intl.formatMessage({
                id: 'Dashboard.components.customModal.ModifyUnitDetailModal.IopsWeight',
                defaultMessage: 'iops权重',
              })}
              name={['unitConfig', 'unitConfig', 'iopsWeight']}
            >
              <InputNumber
                placeholder={intl.formatMessage({
                  id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
                  defaultMessage: '请输入',
                })}
              />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </CustomModal>
  );
}
