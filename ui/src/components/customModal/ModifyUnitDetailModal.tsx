import InputNumber from '@/components/InputNumber';
import { SUFFIX_UNIT,getMinResource } from '@/constants';
import { RULER_ZONE } from '@/constants/rules';
import { TooltipItemContent } from '@/pages/Cluster/New/Observer';
import type {
MaxResourceType,
MinResourceConfig,
} from '@/pages/Tenant/New/ResourcePools';
import ZoneItem from '@/pages/Tenant/ZoneItem';
import {
findMinParameter,
modifyZoneCheckedStatus,
} from '@/pages/Tenant/helper';
import {
createObtenantPoolReportWrap,
patchObtenantPoolReportWrap
} from '@/services/reportRequest/tenantReportReq';
import { formatPatchPoolData } from '@/utils/helper';
import { intl } from '@/utils/intl';
import { useParams } from '@umijs/max';
import { useEffect,useState } from 'react';
import SelectWithTooltip from '../SelectWithTooltip';

import { Col,Form,Row,Select,message } from 'antd';
import type { CommonModalType } from '.';
import CustomModal from '.';

export type PoolDetailType = {
  zoneName?: string; // Exists when adding a new pool
  priority: number;
  selectZone?: string;
  unitConfig: {
    cpuCount: number | string;
    iopsWeight: number;
    logDiskSize: number | string;
    maxIops: number;
    memorySize: number | string;
    minIops: number;
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
  params: {
    clusterList?: API.SimpleClusterList;
    clusterResourceName?: string;
    essentialParameter?: API.EssentialParametersType;
    setClusterList: React.Dispatch<React.SetStateAction<API.SimpleClusterList>>;
    editZone?: string;
    replicaList?: API.ReplicaDetailType[];
    newResourcePool?: boolean;
    setEditZone?: React.Dispatch<React.SetStateAction<string>>;
    zonesOptions?: API.OptionsType;
    zoneName?: string;
  };
};

export default function ModifyUnitDetailModal({
  visible,
  setVisible,
  successCallback,
  params: {
    clusterList = [],
    setClusterList,
    essentialParameter,
    clusterResourceName = '',
    editZone, 
    replicaList,
    newResourcePool = false, // This parameter can be used to determine whether to edit or add
    setEditZone,
    zonesOptions,
    zoneName,
  },
}: CommonModalType & UnitConfigType) {
  const [form] = Form.useForm<PoolDetailType>();
  const { ns, name } = useParams();
  const [maxResource, setMaxResource] = useState<MaxResourceType>({});
  const [minResource, setMinResource] = useState<MinResourceConfig>(
    getMinResource({ minMemory: essentialParameter?.minPoolMemory }),
  );
  
  const [selectZones, setSelectZones] = useState<string[]>(
    editZone ? [editZone] : [],
  );
  const selectZone = Form.useWatch('selectZone', form);
  const obtenantPoolReq = newResourcePool
    ? createObtenantPoolReportWrap
    : patchObtenantPoolReportWrap;

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
    const { zoneName, ...reqData } = formatPatchPoolData(
      values,
      newResourcePool ? 'create' : 'edit',
    );
    const res = await obtenantPoolReq({
      ns,
      name,
      zoneName,
      ...reqData,
    });
    if (res.successful) {
      message.success(
        res.message ||
          intl.formatMessage({
            id: 'Dashboard.components.customModal.ModifyUnitDetailModal.ModifiedSuccessfully',
            defaultMessage: '修改成功',
          }),
      );
      if (successCallback) successCallback();
      form.resetFields();
      if (setEditZone) {
        setEditZone('');
        setSelectZones([]);
      }
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
      cpuCount: zone?.minCPU,
      iopsWeight: zone?.iopsWeight,
      logDiskSize: zone?.logDiskSize.split('Gi')[0],
      maxIops: zone?.maxIops,
      memorySize: zone?.memorySize.split('Gi')[0],
      minIops: zone?.minIops,
    };
    if (newResourcePool) {
      result.priority = zone?.priority;
    } else {
      result[zone?.zone] = {
        priority: zone?.priority,
      };
    }
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
      setMinResource({
        ...minResource,
        minMemory: essentialParameter.minPoolMemory,
      });
    }
  }, [essentialParameter]);

  useEffect(() => {
    if (essentialParameter) {
      if (selectZones.length === 0) {
        setMaxResource({});
        return;
      }
      const maxResource = findMinParameter(selectZones, essentialParameter);
      if (maxResource.maxCPU < minResource.minCPU) {
        maxResource.maxCPU = minResource.minCPU;
      }
      if (maxResource.maxLogDisk < minResource.minLogDisk) {
        maxResource.maxLogDisk = minResource.minLogDisk;
      }
      if (maxResource.maxMemory < minResource.minMemory) {
        maxResource.maxMemory = minResource.minMemory;
      }
      setMaxResource(maxResource);
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

  useEffect(() => {
    if (zoneName && newResourcePool && zonesOptions?.length) {
      form.setFieldValue('zoneName', zoneName);
    }
  }, [zoneName, newResourcePool]);

  return (
    <CustomModal
      width={780}
      title={
        newResourcePool
          ? intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyUnitDetailModal.AddAResourcePool',
              defaultMessage: '新增资源池',
            })
          : intl.formatMessage({
              id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EditResourcePool',
              defaultMessage: '编辑资源池',
            })
      }
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
            <Col span={8}>
              <Form.Item
                name={'zoneName'}
                label={intl.formatMessage({
                  id: 'Dashboard.components.customModal.ModifyUnitDetailModal.ZoneName',
                  defaultMessage: 'Zone名称',
                })}
                validateFirst
                rules={RULER_ZONE}
                style={{ marginRight: 24 }}
              >
                {/* <Input placeholder={'请输入'} /> */}
                <Select
                  onChange={(val: string) => {
                    setSelectZones([val]);
                  }}
                  options={zonesOptions}
                />
              </Form.Item>
            </Col>
            <Col span={4}>
              <Form.Item
                name={'priority'}
                label={intl.formatMessage({
                  id: 'Dashboard.components.customModal.ModifyUnitDetailModal.Weight',
                  defaultMessage: '权重',
                })}
                rules={[
                  {
                    required: true,
                    message: intl.formatMessage({
                      id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnterTheWeightOf',
                      defaultMessage: '请输入所选 zone 的权重',
                    }),
                  },
                ]}
                style={{ marginRight: 24 }}
              >
                <InputNumber />
              </Form.Item>
            </Col>

            <Form.Item
              name={'selectZone'}
              label={intl.formatMessage({
                id: 'Dashboard.components.customModal.ModifyUnitDetailModal.CopyExistingZoneSpecifications',
                defaultMessage: '复制已有zone规格',
              })}
            >
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
                    isEdit={Boolean(editZone)}
                    priorityName={[item.zone, 'priority']}
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
              name={['unitConfig', 'cpuCount']}
              validateFirst
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EnterTheNumberOfCpu',
                    defaultMessage: '请输入 CPU 核数',
                  }),
                },
                () => ({
                  validator() {
                    if (
                      essentialParameter &&
                      findMinParameter(selectZones, essentialParameter).maxCPU <
                        minResource.minCPU
                    ) {
                      return Promise.reject(
                        new Error(
                          intl.formatMessage({
                            id: 'Dashboard.components.customModal.ModifyUnitDetailModal.ZoneCannotCreateAUnit',
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
            >
              <InputNumber
                min={minResource.minCPU}
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
              validateFirst
              name={['unitConfig', 'memorySize']}
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EnterMemory',
                    defaultMessage: '请输入 Memory',
                  }),
                },
                () => ({
                  validator() {
                    if (
                      essentialParameter &&
                      findMinParameter(selectZones, essentialParameter)
                        .maxMemory < minResource.minMemory
                    ) {
                      return Promise.reject(
                        new Error(
                          intl.formatMessage({
                            id: 'Dashboard.components.customModal.ModifyUnitDetailModal.IfTheAvailableMemorySize',
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
            >
              <InputNumber
                min={minResource.minMemory}
                max={maxResource.maxMemory ? maxResource.maxMemory : undefined}
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
              validateFirst
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EnterLogdisksize',
                    defaultMessage: '请输入 logDiskSize',
                  }),
                },
                () => ({
                  validator() {
                    if (
                      essentialParameter &&
                      findMinParameter(selectZones, essentialParameter)
                        .maxLogDisk < minResource.minLogDisk
                    ) {
                      return Promise.reject(
                        new Error(
                          intl.formatMessage({
                            id: 'Dashboard.components.customModal.ModifyUnitDetailModal.ZoneCannotCreateAUnit.1',
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
              label="LogDiskSize"
              name={['unitConfig', 'logDiskSize']}
            >
              <InputNumber
                min={minResource.minLogDisk}
                max={
                  maxResource.maxLogDisk ? maxResource.maxLogDisk : undefined
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
                  rules={[
                    {
                      required: true,
                      message: intl.formatMessage({
                        id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EnterMiniops',
                        defaultMessage: '请输入 minIops',
                      }),
                    },
                  ]}
                  label="min iops"
                  name={['unitConfig', 'minIops']}
                >
                  <InputNumber
                    min={minResource.minIops}
                    placeholder={intl.formatMessage({
                      id: 'Dashboard.components.customModal.ModifyUnitDetailModal.PleaseEnter',
                      defaultMessage: '请输入',
                    })}
                  />
                </Form.Item>
              </Col>
              <Col>
                <Form.Item
                  rules={[
                    {
                      required: true,
                      message: intl.formatMessage({
                        id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EnterMaxiops',
                        defaultMessage: '请输入 maxIops',
                      }),
                    },
                  ]}
                  label="max iops"
                  name={['unitConfig', 'maxIops']}
                >
                  <InputNumber
                    min={minResource.maxIops}
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
              rules={[
                {
                  required: true,
                  message: intl.formatMessage({
                    id: 'Dashboard.components.customModal.ModifyUnitDetailModal.EnterIopsWeight',
                    defaultMessage: '请输入 iops权重',
                  }),
                },
              ]}
              name={['unitConfig', 'iopsWeight']}
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
