import InputNumber from '@/components/InputNumber';
import { SUFFIX_UNIT } from '@/constants';
import { getNSName } from '@/pages/Cluster/Detail/Overview/helper';
import { TooltipItemContent } from '@/pages/Cluster/New/Observer';
import type { MaxResourceType } from '@/pages/Tenant/New/ResourcePools';
import ZoneItem from '@/pages/Tenant/ZoneItem';
import { findMinParameter,modifyZoneCheckedStatus } from '@/pages/Tenant/helper';
import { patchTenantConfiguration } from '@/services/tenant';
import { formatUnitDetailData } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { useEffect,useState } from 'react';
import SelectWithTooltip from '../SelectWithTooltip';

import { Col,Form,Row,message } from 'antd';
import type { CommonModalType } from '.';
import CustomModal from '.';

export type UnitDetailType = {
  unitConfig: {
    unitConfig: {
      cpuCount: number | string;
      iopsWeight: number;
      logDiskSize: number | string;
      maxIops: number;
      memorySize: number | string;
      minIops: number;
    };
    pools: any;
  };
};

const formatReplicaList = (
  replicaList: API.ReplicaDetailType[],
): API.TooltipData[] => {
  return replicaList.map((replica) => ({
    label: replica.zone,
    value: replica.zone,
    toolTipData: Object.keys(replica)
      .filter((key) => key !== 'zone')
      .map((key) => ({ [key]: replica[key] })),
  }));
};

type UnitConfigType = {
  clusterList?: API.SimpleClusterList;
  clusterResourceName?: string;
  essentialParameter?: API.EssentialParametersType;
  setClusterList: React.Dispatch<React.SetStateAction<API.SimpleClusterList>>;
  editZone?: string;
  replicaList?: API.ReplicaDetailType[];
  newResourcePool?: boolean;
  setEditZone?: React.Dispatch<React.SetStateAction<string>>
};

export default function ModifyUnitDetailModal({
  visible,
  setVisible,
  successCallback,
  clusterList = [],
  setClusterList,
  essentialParameter = {},
  clusterResourceName = '',
  editZone,
  replicaList,
  newResourcePool = false,
  setEditZone
}: CommonModalType & UnitConfigType) {
  const [form] = Form.useForm<UnitDetailType>();
  const [minMemory, setMinMemory] = useState<number>(2);
  const [maxResource, setMaxResource] = useState<MaxResourceType>({});
  const [selectZones, setSelectZones] = useState<string[]>(
    editZone ? [editZone] : [],
  );
  const selectZone = Form.useWatch('selectZone', form);
  const handleSubmit = async () => {
    try {
      await form.validateFields();
      form.submit();
    } catch (err) {}
  };
  const handleCancel = () => {
    if (setEditZone) {
      setEditZone('');
      setSelectZones([]);
    }
    form.resetFields();
    setVisible(false);
  };

  const onFinish = async (values: any) => {
    const [ns, name] = getNSName();
    const res = await patchTenantConfiguration({
      ns,
      name,
      ...formatUnitDetailData(values),
    });
    if (res.successful) {
      message.success(res.message || '修改成功');
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
      modifyZoneCheckedStatus(clusterList, name, checked, {
        name: clusterResourceName,
      }),
    );
  };

  const getInitialValues = (editZone: string) => {
    let result = {};
    const zone = replicaList?.find((replica) => replica.zone === editZone);
    result.unitConfig = {
      unitConfig: {
        cpuCount: zone?.minCPU,
        iopsWeight: zone?.iopsWeight,
        logDiskSize: zone?.logDiskSize.split('Gi')[0],
        maxIops: zone?.maxIops,
        memorySize: zone?.memorySize.split('Gi')[0],
        minIops: zone?.minIops,
      },
      pools: {},
    };
    result.unitConfig.pools[editZone] = {
      priority: zone?.priority,
    };

    return result;
  };
  let targetCluster = clusterList.find(
    (cluster) => cluster.name === clusterResourceName,
  );
  let targetZoneList =
    targetCluster?.topology.map((zone) => ({
      zone: zone.zone,
      checked: zone.checked,
    })) || [];

  if (editZone) {
    targetZoneList = targetZoneList.filter((zone) => zone.zone === editZone);
  }

  const selectOptions = formatReplicaList(replicaList || []);
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
  }, [selectZones, essentialParameter]);

  useEffect(() => {
    if (selectZone && replicaList) {
      form.setFieldsValue(getInitialValues(selectZone));
    }
  }, [selectZone]);

  useEffect(() => {
    if (editZone && clusterResourceName) {
      setClusterList(
        modifyZoneCheckedStatus(clusterList, editZone, true, {
          name: clusterResourceName,
        }),
      );
      form.setFieldsValue(getInitialValues(editZone));
      setSelectZones([editZone]);
    }
  }, [editZone]);

  return (
    <CustomModal
      width={780}
      title={newResourcePool ? '新增资源池' : '编辑资源池'}
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
        {newResourcePool && selectOptions.length ? (
          <Row>
            <Form.Item name={'selectZone'} label="复制已有zone规格">
              <SelectWithTooltip
                form={form}
                name={'selectZone'}
                selectList={selectOptions}
                TooltipItemContent={TooltipItemContent}
              />
            </Form.Item>
          </Row>
        ) : (
          <>
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
                    obZoneResource={
                      essentialParameter.obZoneResourceMap[item.zone]
                    }
                    checkBoxOnChange={checkBoxOnChange}
                  />
                ))}
              </Row>
            )}
          </>
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
